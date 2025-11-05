package config

// build: spectrum_exporter

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"time"
	"unisphere_exporter/collectors"

	"gopkg.in/yaml.v3"
)

var cfg *UnisphereConfig

type UnisphereConfig struct {
	Global     *GlobalConfig          `yaml:"global,omitempty"`
	Server     *ServerConfig          `yaml:"server,omitempty"`
	Clients    []*ClientConfig        `yaml:"clients,omitempty"`
	Auths      []*AuthConfig          `yaml:"auths,omitempty"`
	Collectors map[string]interface{} `yaml:"collectors,omitempty"`
}

func NewConfiguration() *UnisphereConfig {
	cfg = &UnisphereConfig{
		Global: &GlobalConfig{
			Server: &GlobalServerConfig{
				Endpoint: "http://127.0.0.1:8080",
				Api_Path: "",
				Insecure: false,
				Mode:     "http",
			},
			Client: &GlobalClientConfig{
				Auth:     "",
				Interval: 1 * time.Minute,
				Insecure: false,
			},
			Provider: &GlobalProviderConfig{},
		},
		Server: &ServerConfig{
			Metrics: &ServerMetricConfig{
				Enabled: true,
			},
			Logs: &ServerLogConfig{
				Enabled: true,
			},
			Traces: &ServerTraceConfig{
				Enabled: true,
			},
		},
		Clients: nil,
		Auths:   nil,
	}
	return cfg
}

func (_cfg *UnisphereConfig) LoadFile(file *string) error {
	ymlContents, err := os.ReadFile(*file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(ymlContents, _cfg)
	if err != nil {
		return err
	}

	err = _cfg.applyGlobal()
	if err != nil {
		return err
	}

	// Check to exist modules in Provider
	for k, _ := range _cfg.Collectors {
		if collectors.RegistryModules[k] == nil {
			return errors.New("provider " + k + " is not registered")
		}
	}

	// Set Configuration Providers...
	for k, v := range collectors.RegistryModules {
		body, err := yaml.Marshal(_cfg.Collectors[k])
		v.Init(k)
		v.SetConfig(body)
		if err != nil {
			return err
		}
	}

	return err
}

// applyGlobal
// Section 내용이 비어있을 경우,
// Global 설정을 각각의 Section에 적용
func (_cfg *UnisphereConfig) applyGlobal() error {
	// Set Client
	g := _cfg.Global
	if _cfg.Clients == nil {
		return errors.New("no clients configured")
	}
	for _, c := range _cfg.Clients {
		if c.Endpoint == "" {
			return errors.New("client endpoint is required")
		}
		if c.Auth == "" {
			c.Auth = g.Client.Auth
		}
		if c.Insecure == "" {
			c.Insecure = strconv.FormatBool(g.Client.Insecure)
		}
		if c.Interval == "" {
			c.Interval = g.Client.Interval.String()
		}
		for k, v := range g.Client.Labels {
			if c.Labels[k] == "" {
				c.Labels[k] = v
			}
		}
	}
	// Set Global config at Servers
	svNum := reflect.ValueOf(_cfg.Server).Elem().NumField()
	for i := 0; i < svNum; i++ {
		sv := reflect.ValueOf(_cfg.Server).Elem().Field(i).Elem()
		endpoint := sv.FieldByName("Endpoint")
		if endpoint.String() == "" {
			endpoint.SetString(g.Server.Endpoint)
		}
		apiPath := sv.FieldByName("Api_Path")
		if apiPath.String() == "" {
			apiPath.SetString(g.Server.Api_Path)
		}
		insecure := sv.FieldByName("Insecure")
		if insecure.String() == "" {
			insecure.SetString(strconv.FormatBool(g.Server.Insecure))
		}
		mode := sv.FieldByName("Mode")
		if mode.String() == "" {
			mode.SetString(g.Server.Mode)
		}
	}

	return nil
}

// SearchAuth
// 인증정보를 찾아, base64로 인코딩하여 리턴합니다.
func (_cfg *UnisphereConfig) SearchAuth(name string) (string, string) {
	for _, auth := range _cfg.Auths {
		if auth.Name == name {
			return auth.User, auth.Password
		}
	}
	return "", ""
}
