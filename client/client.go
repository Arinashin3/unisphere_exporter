package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type Targets struct {
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	SkipSsl   bool   `yaml:"skip_ssl,omitempty"`
	SslVerify bool   `yaml:"ssl_verify,omitempty"`
	Cert      string `yaml:"cert,omitempty"`
	Timeout   string `yaml:"timeout,omitempty"`
	loaded    bool   `yaml:"loaded,omitempty"`
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
	Modules map[string]Targets `yaml:"modules"`
}

var (
	configMap Configs
	roots     *x509.CertPool
)

func NewClient(tgt string, mod string, logger *slog.Logger) (*UnisphereClient, bool) {
	var uc UnisphereClient
	uc.url.Host = tgt
	uc.Logger = logger
	if !uc.searchModule(mod) {
		return &uc, false
	}
	return &uc, uc.tryLogin()
}

func (uc *UnisphereClient) tryLogin() bool {
	//uc.url.Scheme = "http"
	tgt := uc.url
	tgt.Path = "/api/types/user/instances"
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

// SetModules will Read Config file's module
func SetModules(cfgFile *string, logger *slog.Logger) bool {
	var result bool
	cfg, err := os.ReadFile(*cfgFile)
	cfgMap := &configMap
	cfgMap.Modules = make(map[string]Targets)
	if err != nil {
		logger.Error("Failed to read Config File: %v", cfgFile)
	}
	if yaml.Unmarshal(cfg, &cfgMap) != nil {
		logger.Error("Failed to Unmarshal Config File: %v", err)
		return result
	}

	roots, err = x509.SystemCertPool()
	if err != nil {
		logger.Error("Unable to fetch system CA store.")
		return result
	}

	for k, v := range cfgMap.Modules {
		if v.Cert != "" {
			certs, err := os.ReadFile(v.Cert)
			if err != nil {
				logger.Error("Failed to read extra CA file.", "module", k)
				continue
			}
			if !roots.AppendCertsFromPEM(certs) {
				logger.Error("Failed to append certs from PEM, unknown error.", "module", k)
				continue
			}
		}
		if v.Timeout == "" {
			v.Timeout = "10s"
		}
		v.loaded = true
		cfgMap.Modules[k] = v
	}
	logger.Info("Loaded Credentials Modules", "api_count", len(cfgMap.Modules))
	result = true
	return result
}

// The getModule function fetches authentication information with keys that match the module.
func (uc *UnisphereClient) getModule(cfg Targets) bool {
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
		return result
	}
	uc.hc = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tc},
		Timeout:   to,
	}

	result = true
	return result
}

func (uc *UnisphereClient) searchModule(module string) bool {
	cfg := configMap.Modules[module]
	if !cfg.loaded {
		uc.Logger.Error("Unknown Module", "module", module)
		return false
	}

	return uc.getModule(cfg)
}

func (uc *UnisphereClient) Get(path string, query string) []byte {
	tgt := uc.url
	tgt.Path = path
	tgt.RawQuery = query
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
