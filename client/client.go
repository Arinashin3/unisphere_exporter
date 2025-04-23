package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type Targets struct {
	ModuleName string `yaml:"module_name"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	SkipSsl    bool   `yaml:"skip_ssl,omitempty"`
	SslVerify  bool   `yaml:"ssl_verify,omitempty"`
	Cert       string `yaml:"cert,omitempty"`
	Timeout    string `yaml:"timeout,omitempty"`
}

type UnisphereClient struct {
	url    url.URL
	auth   string
	ctx    context.Context
	token  string
	Logger *slog.Logger
	hc     *http.Client
}

type Modules struct {
	ModuleName interface{}
}

type Configs struct {
	Modules []Targets `yaml:"modules"`
}

var (
	cfgMap Configs
	roots  *x509.CertPool
)

func NewClient(tgt string, mod string, logger *slog.Logger) (*UnisphereClient, bool) {
	var uc UnisphereClient
	uc.url.Host = tgt
	uc.Logger = logger
	uc.searchModule(mod)
	result := uc.tryLogin()

	return &uc, result
}

func (uc *UnisphereClient) tryLogin() bool {
	tgt := uc.url
	tgt.Path = "/api/types/user/instances"
	tgt.Scheme = "http"
	req, err := http.NewRequest("GET", tgt.String(), nil)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
	}
	uc.hc.Jar, _ = cookiejar.New(nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")

	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
		return false
	}
	defer resp.Body.Close()
	uc.token = resp.Header.Get("Emc-Csrf-Token")
	return true
}

func SetModule(cfgFile *string, logger *slog.Logger) {
	cfg, err := os.ReadFile(*cfgFile)
	if err != nil {
		logger.Error("Failed to read Config File: %v", cfgFile)
	}
	if yaml.Unmarshal(cfg, &cfgMap) != nil {
		log.Fatalf("Failed to Unmarshal Config File: %v", err)
	}

	roots, err = x509.SystemCertPool()
	if err != nil {
		logger.Error("Unable to fetch system CA store.")
	}

	for i, v := range cfgMap.Modules {
		if v.Cert != "" {
			certs, err := os.ReadFile(v.Cert)
			if err != nil {
				logger.Error("Failed to read extra CA file.", "cert_file", v.Cert)
			}
			if !roots.AppendCertsFromPEM(certs) {
				logger.Error("Failed to append certs from PEM, unknown error.", "error", err)
			}

		}
		if v.Timeout == "" {
			cfgMap.Modules[i].Timeout = "10s"
		}

	}
	logger.Info("Loaded API Credentials", "api_count", len(cfgMap.Modules))
}

func (uc *UnisphereClient) getConfig(cfg Targets) bool {
	var result bool
	uc.auth = base64.StdEncoding.EncodeToString([]byte(cfg.User + ":" + cfg.Password))

	if cfg.SkipSsl {
		uc.url.Scheme = "http"
	} else {
		uc.url.Scheme = "https"
	}
	tc := &tls.Config{RootCAs: roots}
	tc.InsecureSkipVerify = !cfg.SslVerify
	to, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		uc.Logger.Error("Failed Parse the timeout", "timeduration", cfg.Timeout)
	}
	uc.hc = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tc},
		Timeout:   to,
	}

	return result
}

func (uc *UnisphereClient) searchModule(module string) bool {
	for _, v := range cfgMap.Modules {
		if v.ModuleName == module {
			return uc.getConfig(v)
		}
	}
	uc.Logger.Error("Failed Search Module at Config File.", "module", module)
	return false
}

func (uc *UnisphereClient) Get(path string, query string) []byte {
	tgt := uc.url
	tgt.Path = path
	tgt.RawQuery = query
	tgt.Scheme = "http"
	req, err := http.NewRequest("GET", tgt.String(), nil)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	//req.WithContext(uc.ctx)

	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Debug("Failed to request", "path", path, "err", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.Logger.Error("Failed to read body", "path", path, "err", err)
		return nil
	}
	return body
}

func (uc *UnisphereClient) GetMetricQuery(path string, query string) []byte {
	tgt := uc.url
	tgt.Path = path
	tgt.RawQuery = query
	tgt.Scheme = "http"
	req, err := http.NewRequest("GET", tgt.String(), nil)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	//req.WithContext(uc.ctx)

	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Debug("Failed to request", "path", path, "err", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.Logger.Error("Failed to read body", "path", path, "err", err)
		return nil
	}
	return body
}
