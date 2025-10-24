package config

// build: spectrum_exporter

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
	"unisphere_exporter/provider"

	"gopkg.in/yaml.v3"
)

var cfg *UnisphereConfig

type UnisphereConfig struct {
	Global    *GlobalConfig              `yaml: "global,omitempty"`
	Server    *ServerConfig              `yaml: "server,omitempty"`
	Clients   []*ClientConfig            `yaml: "targets,omitempty"`
	Auths     []*AuthConfig              `yaml: "auths,omitempty"`
	Providers map[string]provider.Option `yaml: "providers,omitempty"`
}

type Providers struct {
	System   *ProviderDefaults `yaml:"system,omitempty"`
	Lun      *ProviderDefaults `yaml:"lun,omitempty"`
	Capacity *ProviderDefaults `yaml:"capacity,omitempty"`
	Metric_A *ProviderMetric   `yaml:"metric_a,omitempty"`
	Metric_B *ProviderMetric   `yaml:"metric_b,omitempty"`
	Metric_C *ProviderMetric   `yaml:"metric_c,omitempty"`
	Event    *ProviderEvent    `yaml:"event,omitempty"`
	Alert    *ProviderEvent    `yaml:"alert,omitempty"`
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
		//Providers: make(map[string]interface{}),
		Providers: provider.ModuleOptions,
	}
	return cfg
}

func (_cfg *UnisphereConfig) LoadFile(file *string) error {
	ymlContents, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	tmp := _cfg.Providers

	err = yaml.Unmarshal(ymlContents, _cfg)
	for k, v := range tmp {
		fmt.Println(k, v)
	}
	if err != nil {
		return err
	}

	err = _cfg.applyGlobal()
	if err != nil {
		return err
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

func GetConfig() *UnisphereConfig {
	return cfg
}

func (_cfg *UnisphereConfig) GetMetricsEndpoint() string {
	if _cfg.Server.Metrics.Enabled {
		return _cfg.Server.Metrics.Endpoint + _cfg.Server.Metrics.Api_Path
	}
	return ""
}

func (_cfg *UnisphereConfig) GetMetricsMode() string {
	if _cfg.Server.Metrics.Enabled {
		return _cfg.Server.Metrics.Mode
	}
	return ""
}

func (_cfg *UnisphereConfig) GetMetricsInsecure() bool {
	insecure, _ := strconv.ParseBool(_cfg.Server.Metrics.Insecure)
	return insecure
}

func (_cfg *UnisphereConfig) GetLogsEndpoint() string {
	if _cfg.Server.Logs.Enabled {
		return _cfg.Server.Logs.Endpoint + _cfg.Server.Logs.Api_Path
	}
	return ""
}
func (_cfg *UnisphereConfig) GetLogsMode() string {
	if _cfg.Server.Logs.Enabled {
		return _cfg.Server.Logs.Mode
	}
	return ""
}

func (_cfg *UnisphereConfig) GetLogsInsecure() bool {
	insecure, _ := strconv.ParseBool(_cfg.Server.Logs.Insecure)
	return insecure
}

func (_cfg *UnisphereConfig) GetClientList() []*ClientConfig {
	return _cfg.Clients
}
