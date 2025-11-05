package collectors

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.yaml.in/yaml/v3"
)

func init() {
	key := "basicSystemInfo"
	RegisterModule(key, NewBasicSystemInfo())
}

type ModuleBasicSystemInfo struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool // Default Enabled

	// Configuration File
	Enabled *bool `yaml:"enabled"`
}

func NewBasicSystemInfo() *ModuleBasicSystemInfo {
	return &ModuleBasicSystemInfo{
		defaults: true,
	}
}

func (_m *ModuleBasicSystemInfo) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "info",
			Name:     "unisphere_basic_system_info",
			Desc:     "Information about unisphere basicSystemInfo",
			Unit:     "",
			TypeName: "gauge",
		},
	}
	_m.opts = api.NewUnityActionOptions("basicSystemInfo")
	_m.opts.Fields = []string{"model", "softwareFullVersion"}
}

func (_m *ModuleBasicSystemInfo) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleBasicSystemInfo) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleBasicSystemInfo) Run(logger *slog.Logger, col *Collector) {
	meter := col.meterProvider.Meter(_m.name)
	client := col.Client

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, _m.desc, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// Set Attributes
		if col.labels == nil {
			logger.Debug("hostLabels not set")
			return nil
		}
		clientAttrs := metric.WithAttributes(append(col.resource.Attributes(), col.labels...)...)

		// Request Data
		data, err := client.GetInstances(_m.opts)
		if err != nil {
			logger.Error("Failed to get", "error", err, "module", _m.name)
			return nil
		}

		// System Attributes...
		for _, v := range data {
			infoAttrs := metric.WithAttributes(attribute.String("product.name", v.Get("model").String()), attribute.String("firmware.version", v.Get("softwareFullVersion").String()))
			observer.ObserveFloat64(observableMap["info"], 1, clientAttrs, infoAttrs)
		}

		return nil
	}, observableArray...)

}
