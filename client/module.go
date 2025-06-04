package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Module struct {
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	SkipSsl   bool   `yaml:"skip_ssl,omitempty"`
	SslVerify bool   `yaml:"ssl_verify,omitempty"`
	Cert      string `yaml:"cert,omitempty"`
	Timeout   string `yaml:"timeout,omitempty"`
	loaded    bool   `yaml:"loaded,omitempty"`
}

type ModuleMap struct {
	Modules map[string]Module `yaml:"modules"`
}

var (
	moduleList map[string]Module
	roots      *x509.CertPool
)

func InitModules(fileName *string, logger *slog.Logger) bool {
	moduleList = make(map[string]Module)
	contents, err := os.ReadFile(*fileName)
	if err != nil {
		logger.Error("Failed to read Config File", "file_name", fileName)
	}

	var config ModuleMap
	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		logger.Error("Unmarchal Error", "error", err)
	}

	for k, v := range config.Modules {
		if v.Cert != "" {
			var certs []byte
			certs, err = os.ReadFile(v.Cert)
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
		moduleList[k] = v

	}

	logger.Info("Number of registered modules.", "count", len(config.Modules))

	return true
}

// The getModule function fetches authentication information with keys that match the module.
func (uc *UnisphereClient) getModule(modName string) bool {
	var result bool
	cfg := moduleList[modName]
	logger := uc.Logger
	if !cfg.loaded {
		logger.Error("Unknown Module", "module", modName)
		return false
	}
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

func searchModule(module string, logger *slog.Logger) Module {
	mod := moduleList[module]

	if !mod.loaded {
		return moduleList[module]
	}

	return moduleList[module]
}
