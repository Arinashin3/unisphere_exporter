package config

import (
	"strconv"
	"time"
)

type ProviderMetric struct {
	Enabled  string   `yaml:"enabled,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
}

func (pv *ProviderMetric) GetEnabled(defaults bool) bool {
	if pv.Enabled == "" {
		return defaults
	}
	enabled, _ := strconv.ParseBool(pv.Enabled)
	return enabled
}

func (pv *ProviderMetric) GetInterval() time.Duration {
	interval, _ := time.ParseDuration(pv.Interval)
	return interval
}
