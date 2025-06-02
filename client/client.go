package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type UnisphereClient struct {
	url           url.URL
	auth          string
	ctx           context.Context
	Logger        *slog.Logger
	hc            *http.Client
	loginAt       time.Time
	lastConnectAt time.Time
	token         string
}

var clientList map[string]*UnisphereClient

func init() {
	clientList = make(map[string]*UnisphereClient)
}

func ReadyClient(target string, mod string, logger *slog.Logger) *UnisphereClient {
	loginDuration := time.Now().AddDate(0, 0, -1).Unix()
	if clientList[target] == nil {
		clientList[target] = newClient(target, mod, logger)
	}
	if clientList[target].loginAt.Unix() < loginDuration {
		logger.Info("Expire and recreate client.", "client", target)
		clientList[target] = expireClient(target, logger)
		clientList[target] = newClient(target, mod, logger)
	}
	uc := clientList[target]
	respBody := uc.Send("GET", "/api/types/loginSessionInfo/instances", "compact=true", nil)
	if respBody == nil {
		logger.Debug("Client is not ready.", "client", target)
		return nil
	} else {
		logger.Debug("Client is ready.", "client", target)
		clientList[target].lastConnectAt = time.Now()
	}

	return clientList[target]
}

func (uc *UnisphereClient) Send(method string, path string, query string, body io.Reader) []byte {
	logger := uc.Logger
	u := uc.url
	u.Path = path
	u.RawQuery = query
	logger.Debug("Create Request", "method", method)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		logger.Error("Failed to Create New Request.", "error", err)
		return nil
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.WithContext(uc.ctx)

	// Header for Login and Set CookieJar
	if uc.hc.Jar == nil {
		logger.Debug("Cannot find the Cookies.", "client", u.Host)
		logger.Info("Login to use Basic Auth.", "client", u.Host)
		logger.Debug("Create New Cookie.", "client", u.Host)
		req.Header.Add("X-EMC-REST-CLIENT", "true")
		req.Header.Add("Authorization", "Basic "+uc.auth)
		uc.hc.Jar, err = cookiejar.New(nil)
		if err != nil {
			logger.Error("Failed to Create CookieJar.", "url", u.String(), "error", err)
			return nil
		}
	}

	// Header for POST Request
	if method == "POST" {
		req.Header.Add("EMC-CSRF-TOKEN", uc.token)
	}

	resp, err := uc.hc.Do(req)
	if err != nil {
		logger.Error("Response Error", "error", err)
		return nil
	}
	if int(resp.StatusCode/100) != 2 {
		logger.Error("Http Code is Not 2xx.", "http_code", resp.StatusCode)
		return nil
	}

	if uc.token == "" {
		uc.token = resp.Header.Get("Emc-Csrf-Token")
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return respBody
}

func newClient(tgt string, mod string, logger *slog.Logger) *UnisphereClient {
	var uc UnisphereClient
	uc.url.Host = tgt
	uc.Logger = logger
	uc.ctx = context.Background()
	if !uc.searchModule(mod) {
		logger.Error("Cannot find module", "module", mod)
		return nil
	}
	resp := uc.Send("GET", "/api/types/loginSessionInfo/instances", "", nil)
	if resp == nil {
		return nil
	}
	uc.loginAt = time.Now()
	logger.Info("Get New Token.", "client", uc.url.Host)

	return &uc
}

func expireClient(tgt string, logger *slog.Logger) *UnisphereClient {
	resp := clientList[tgt].Send("POST", "/api/types/loginSessionInfo/action/logout", "", nil)
	logger.Info("rep", "rep", resp)
	return nil

}

// SetModules will Read Config file's module

func (uc *UnisphereClient) Get(path string, query string) []byte {
	tgt := uc.url
	tgt.Path = path
	tgt.RawQuery = query

	// New Requset
	body := uc.Send("GET", path, query, nil)

	return body
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
