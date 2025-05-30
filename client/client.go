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
	url           url.URL
	auth          string
	ctx           context.Context
	Logger        *slog.Logger
	hc            *http.Client
	loginAt       time.Time
	lastConnected time.Time
	token         string
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
	ucList    map[string]*UnisphereClient
)

func init() {
	ucList = make(map[string]*UnisphereClient)
}

func GetClient(tgt string, mod string, logger *slog.Logger) {
	u := ucList[tgt]
	loginDuration := time.Now().AddDate(0, 0, -1).Second()
	if u == nil {
		CreateClient(tgt, mod, logger)
		ucList[tgt].ClientLogIn()
	}
	if u.loginAt.Second() < loginDuration {
		ucList[tgt].ClientLogout()
	}

	ucList[tgt].lastConnected = time.Now()
}

func (uc *UnisphereClient) ClientLogIn() bool {
	var result bool
	var err error
	logger := uc.Logger

	uc.url.Path = "/api/types/loginSessionInfo/instances"
	uc.hc.Jar, err = cookiejar.New(nil)
	if err != nil {
		logger.Error("Failed to Create CookieJar.", "url", uc.url.String(), "error", err)
	}

	// Create Requert
	req, err := http.NewRequest("GET", uc.url.String(), nil)
	if err != nil {
		logger.Error("Failed to Create Request.", "error", err)
		return result
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.WithContext(uc.ctx)

	var resp *http.Response
	resp, err = uc.hc.Do(req)
	if err != nil {
		logger.Error("Response Error", "error", err)
		return false
	}
	if int(resp.StatusCode/100) != 2 {
		logger.Error("Http Code is Not 2xx.", "http_code", resp.StatusCode)
		return false
	}

	uc.loginAt = time.Now()
	uc.token = resp.Header.Get("Emc-Csrf-Token")
	return true
}

func (uc *UnisphereClient) ClientLogout() bool {
	var result bool
	var err error
	logger := uc.Logger

	uc.url.Path = "/api/types/loginSessionInfo/action/logout"

	// Create Request
	req, err := http.NewRequest("POST", uc.url.String(), nil)
	if err != nil {
		logger.Error("Failed to Create Request.", "error", err)
		return result
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("EMC-CSRF-TOKEN", uc.token)
	//req.Header.Add("X-EMC-REST-CLIENT", "true")
	//req.Header.Add("Authorization", "Basic "+uc.auth)
	req.WithContext(uc.ctx)

	var resp *http.Response
	resp, err = uc.hc.Do(req)
	if err != nil {
		logger.Error("Response Error", "error", err)

		return false

	}
	if int(resp.StatusCode/100) != 2 {
		logger.Error("Http Code is Not 2xx.", "http_code", resp.StatusCode)
		return false
	}
	return true
}

func CreateClient(tgt string, mod string, logger *slog.Logger) (*UnisphereClient, bool) {
	var uc UnisphereClient
	uc.url.Host = tgt
	uc.Logger = logger
	uc.ctx = context.Background()
	if !uc.searchModule(mod) {
		return &uc, false
	}
	return &uc, uc.tryLogin()
}

func (uc *UnisphereClient) tryLogin() bool {
	//uc.url.Scheme = "http"
	tgt := uc.url
	tgt.Path = "/api/types/loginSessionInfo/instances"
	ctx, cancel := context.WithTimeout(uc.ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", tgt.String(), nil)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
	}
	uc.hc.Jar, _ = cookiejar.New(nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.WithContext(ctx)

	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
		return false
	}
	defer resp.Body.Close()

	//uc.token = resp.Header.Get("Emc-Csrf-Token")
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
	ctx, cancel := context.WithTimeout(uc.ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", tgt.String(), nil)
	if err != nil {
		uc.Logger.Error("Login Failed", "error", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.WithContext(ctx)

	resp, err := uc.hc.Do(req)
	defer resp.Body.Close()
	if err != nil {
		uc.Logger.Debug("Failed to request", "path", path, "err", err)
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.Logger.Error("Failed to read body", "path", path, "err", err)
		return nil
	}
	return body
}
