package config

import (
	"strconv"
	"time"
)

type ProviderEvent struct {
	Enabled  string `yaml:"enabled,omitempty"`
	Level    int    `yaml:"level,omitempty"`
	Interval string `yaml:"interval,omitempty"`
}

func (pv *ProviderEvent) GetEnabled(defaults bool) bool {
	if pv.Enabled == "" {
		return defaults
	}
	enabled, _ := strconv.ParseBool(pv.Enabled)
	return enabled
}

func (pv *ProviderEvent) GetInterval() time.Duration {
	interval, _ := time.ParseDuration(pv.Interval)
	return interval
}
