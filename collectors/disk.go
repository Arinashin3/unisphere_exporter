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
	key := "disk"
	RegisterModule(key, NewDisk())
}

type ModuleDisk struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool
	labels   []string

	// Configuration File
	Enabled *bool
}

func NewDisk() *ModuleDisk {
	return &ModuleDisk{
		defaults: false,
	}
}

func (_m *ModuleDisk) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "info",
			Name:     "unisphere_disk_info",
			Desc:     "information of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		{
			Key:      "health.value",
			Name:     "unisphere_disk_health",
			Desc:     "Health of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		{
			Key:      "size",
			Name:     "unisphere_disk_size",
			Desc:     "Usable capacity",
			Unit:     "mb",
			TypeName: "gauge",
		},
		{
			Key:      "isInUse",
			Name:     "unisphere_disk_is_in_use",
			Desc:     "Indicates whether the drive contains user-written data",
			Unit:     "",
			TypeName: "gauge",
		},
	}
	_m.labels = []string{"id", "name", "model", "emcPartNumber", "slotNumber"}
	_m.opts = api.NewUnityActionOptions("disk")
	for _, v := range _m.desc {
		if v.Key == "info" {
			continue
		}
		_m.opts.Fields = append(_m.opts.Fields, v.Key)
	}
	_m.opts.Fields = append(_m.opts.Fields, _m.labels...)
}

func (_m *ModuleDisk) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleDisk) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleDisk) Run(logger *slog.Logger, col *Collector) {
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
			diskAttrs := metric.WithAttributes(
				attribute.String("disk.id", v.Get("id").String()),
				attribute.String("slot.id", v.Get("slotNumber").String()),
			)
			infoAttrs := metric.WithAttributes(
				attribute.String("disk.model", v.Get("model").String()),
				attribute.String("disk.part", v.Get("emcPartNumber").String()),
			)
			for _, desc := range _m.desc {
				key := desc.Key
				switch key {
				case "info":
					var f float64
					if v.Get("emcPartNumber").String() != "" {
						f = 1
					}
					observer.ObserveFloat64(observableMap[key], f, clientAttrs, diskAttrs, infoAttrs)
					continue
				case "health.value":
					observer.ObserveFloat64(observableMap[key], v.Get(key).Float(), clientAttrs, diskAttrs)
				case "size":
					observer.ObserveFloat64(observableMap[key], utils.Bytes(v.Get(key).Int()).ToMiB(), clientAttrs, diskAttrs)
				case "isInUse":
					var f float64
					if v.Get(key).Bool() {
						f = 1
					}
					observer.ObserveFloat64(observableMap[key], f, clientAttrs, diskAttrs)
				}
			}
		}

		return nil
	}, observableArray...)

}
