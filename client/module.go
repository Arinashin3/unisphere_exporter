package client

import (
	"crypto/x509"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
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
