package client

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type UnisphereClient struct {
	url        url.URL
	moduleName string
	auth       string
	ctx        context.Context
	Logger     *slog.Logger
	hc         *http.Client
	connection struct {
		lastConnect   bool
		lastConnectAt time.Time
		loginAt       time.Time
	}
	token string
}

var clientList map[string]*UnisphereClient

func init() {
	clientList = make(map[string]*UnisphereClient)
}

func LoadClient(target string, module string, logger *slog.Logger) *UnisphereClient {
	if !CheckClient(target, module, logger) {
		clientList[target] = newClient(target, module, logger)
		if clientList[target] == nil {
			return nil
		}
	}

	respBody := clientList[target].Send("GET", "/api/types/loginSessionInfo/instances", "compact=true", nil)
	if respBody == nil {
		logger.Debug("cannot load client", "client", target, "module", module)
	}

	logger.Debug("client loaded", "client", target)
	clientList[target].connection.lastConnectAt = time.Now()

	return clientList[target]
}

func CheckClient(target string, mod string, logger *slog.Logger) bool {
	loginDuration := time.Now().AddDate(0, 0, -1).Unix()

	if clientList[target] == nil {
		logger.Debug("the client is not setting yet", "client", target, "module", mod)
		return false
	}
	if !clientList[target].connection.lastConnect {
		logger.Info("Last connection was failed, recreate client", "client", target, "module", mod)
		clientList[target] = expireClient(target, logger)
		delete(clientList, target)
		return false
	}
	if clientList[target].moduleName != mod {
		logger.Info("Use other module", "client", target, "module", mod)
		clientList[target] = expireClient(target, logger)
		delete(clientList, target)
		return false
	}
	if clientList[target].connection.loginAt.Unix() < loginDuration {
		logger.Info("The login time is too old.", "client", target, "module", mod)
		clientList[target] = expireClient(target, logger)
		delete(clientList, target)
		return false
	}

	return true
}

func (uc *UnisphereClient) Send(method string, path string, query string, body io.Reader) []byte {
	logger := uc.Logger
	u := uc.url
	u.Path = path
	u.RawQuery = query
	logger.Debug("Create Request", "method", method, "path", path, "query", query)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		logger.Error("Failed to Create New Request.", "error", err)
		uc.connection.lastConnect = false
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
			uc.connection.lastConnect = false
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
		uc.connection.lastConnect = false
		return nil
	}
	if int(resp.StatusCode/100) != 2 {
		logger.Error("Http Code is Not 2xx.", "http_code", resp.StatusCode)
		uc.connection.lastConnect = false
		return nil
	}

	if uc.token == "" {
		logger.Debug("Login Success", "client", uc.url.Host)
		uc.token = resp.Header.Get("Emc-Csrf-Token")
		uc.connection.loginAt = time.Now()
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Response Error", "error", err)
		uc.connection.lastConnect = false
		return nil
	}
	uc.connection.lastConnectAt = time.Now()
	uc.connection.lastConnect = true

	return respBody
}

func newClient(tgt string, modName string, logger *slog.Logger) *UnisphereClient {
	var uc UnisphereClient
	uc.url.Host = tgt
	uc.Logger = logger
	uc.ctx = context.Background()
	uc.moduleName = modName

	uc.getModule(modName)

	if !searchModule(modName, logger).loaded {
		return nil
	}
	resp := uc.Send("GET", "/api/types/loginSessionInfo/instances", "", nil)
	if resp == nil {
		return nil
	}
	uc.connection.loginAt = time.Now()
	logger.Info("Get New Token.", "client", uc.url.Host)

	return &uc
}

func expireClient(target string, logger *slog.Logger) *UnisphereClient {
	clientList[target].Send("POST", "/api/types/loginSessionInfo/action/logout", "", nil)
	logger.Debug("Logout Session", "client", target)

	return nil
}
