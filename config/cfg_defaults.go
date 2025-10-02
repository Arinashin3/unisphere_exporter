package config

import (
	"strconv"
	"time"
)

type ProviderDefaults struct {
	Enabled  string `yaml:"enabled,omitempty"`
	Interval string `yaml:"interval,omitempty"`
}

func (pv *ProviderDefaults) GetEnabled(defaults bool) bool {
	if pv.Enabled == "" {
		return defaults
	}
	enabled, _ := strconv.ParseBool(pv.Enabled)
	return enabled
}

func (pv *ProviderDefaults) GetInterval() time.Duration {
	interval, _ := time.ParseDuration(pv.Interval)
	return interval
}
