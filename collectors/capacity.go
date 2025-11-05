package collectors

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils"

	"go.opentelemetry.io/otel/metric"
	"gopkg.in/yaml.v3"
)

type ModuleSystemCapacity struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool

	// Configuration File
	Enabled *bool `yaml:"enabled"`
}

func init() {
	key := "systemCapacity"
	RegisterModule(key, NewSystemCapacity())
}

func NewSystemCapacity() *ModuleSystemCapacity {
	return &ModuleSystemCapacity{
		defaults: true,
	}
}

func (_m *ModuleSystemCapacity) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "sizeTotal",
			Name:     "unisphere_capacity_total_capacity",
			Desc:     "Total capacity of unisphere capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizeUsed",
			Name:     "unisphere_capacity_used_capacity",
			Desc:     "Used capacity of unisphere capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizeFree",
			Name:     "unisphere_capacity_free_capacity",
			Desc:     "Free capacity of unisphere capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizePreallocated",
			Name:     "unisphere_capacity_preallocated_capacity",
			Desc:     "Total provisioned capacity of unisphere capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "totalLogicalSize",
			Name:     "unisphere_capacity_total_provision",
			Desc:     "Total provisioned capacity of unisphere capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
	}
	_m.opts = api.NewUnityActionOptions("systemCapacity")
	for _, desc := range _m.desc {
		_m.opts.Fields = append(_m.opts.Fields, desc.Key)
	}
}

func (_m *ModuleSystemCapacity) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleSystemCapacity) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleSystemCapacity) Run(logger *slog.Logger, col *Collector) {
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
			col.success = false
			return nil
		}
		col.success = true

		// Capacity Attributes...
		for _, v := range data {
			for _, desc := range _m.desc {
				key := desc.Key
				observer.ObserveFloat64(observableMap[key], utils.Bytes(v.Get(key).Int()).ToMiB(), clientAttrs)
			}
		}

		return nil
	}, observableArray...)

}
