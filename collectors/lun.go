package collectors

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils"

	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/otel/metric"
)

func init() {
	key := "lun"
	RegisterModule(key, NewLun())
}

type ModuleLun struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool

	// Configuration File
	Enabled    *bool
	ExcludeLun []string
}

func NewLun() *ModuleLun {
	return &ModuleLun{
		defaults: false,
	}
}

func (_m *ModuleLun) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "sizeTotal",
			Name:     "unisphere_lun_total_size",
			Desc:     "Total Size lun of unisphere",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizeUsed",
			Name:     "unisphere_lun_used_size",
			Desc:     "Used Size lun of unisphere",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizeAllocated",
			Name:     "unisphere_lun_allocated_size",
			Desc:     "Size of space actually allocated in the pool for the LUN.",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "sizePreallocated",
			Name:     "unisphere_lun_preallocated_size",
			Desc:     "Total provisioned lun of unisphere lun",
			Unit:     "mb",
			TypeName: "gauge",
		},
	}
	_m.opts = api.NewUnityActionOptions("lun")
	for _, v := range _m.desc {
		_m.opts.Fields = append(_m.opts.Fields, v.Key)
	}
	_m.opts.Fields = append(_m.opts.Fields, "name", "id")
}

func (_m *ModuleLun) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleLun) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_pv *ModuleLun) Run(logger *slog.Logger, col *Collector) {
	meter := col.meterProvider.Meter(_pv.name)
	client := col.Client

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, _pv.desc, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// Set Attributes
		if col.labels == nil {
			return nil
		}
		clientAttrs := metric.WithAttributes(append(col.resource.Attributes(), col.labels...)...)

		// Request Data
		data, err := client.GetInstances(_pv.opts)
		if err != nil {
			logger.Error("Failed to get", "error", err, "module", _pv.name)
			col.success = false
			return nil
		}
		col.success = true

		// Capacity Attributes...
		for _, v := range data {
			lunAttrs := metric.WithAttributes(attribute.String("lun.id", v.Get("id").String()), attribute.String("lun.name", v.Get("name").String()))
			for _, desc := range _pv.desc {
				key := desc.Key
				observer.ObserveFloat64(observableMap[key], utils.Bytes(v.Get(key).Int()).ToMiB(), clientAttrs, lunAttrs)
			}
		}

		return nil
	}, observableArray...)

}
