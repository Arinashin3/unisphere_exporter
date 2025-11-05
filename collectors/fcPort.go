package collectors

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"

	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/otel/metric"
)

func init() {
	key := "fcPort"
	RegisterModule(key, NewFcPort())
}

type ModuleFcPort struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool
	labels   []string

	// Configuration File
	Enabled *bool
}

func NewFcPort() *ModuleFcPort {
	return &ModuleFcPort{
		defaults: false,
	}
}

func (_m *ModuleFcPort) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "info",
			Name:     "unisphere_fcPort_info",
			Desc:     "information of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		{
			Key:      "health.value",
			Name:     "unisphere_fcPort_health",
			Desc:     "Health of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		{
			Key:      "currentSpeed",
			Name:     "unisphere_fcPort_current_speed",
			Desc:     "Usable capacity",
			Unit:     "mbps",
			TypeName: "gauge",
		},
	}
	_m.labels = []string{"id", "name", "wwn", "slotNumber"}
	_m.opts = api.NewUnityActionOptions("fcPort")
	for _, v := range _m.desc {
		if v.Key == "info" {
			continue
		}
		_m.opts.Fields = append(_m.opts.Fields, v.Key)
	}
	_m.opts.Fields = append(_m.opts.Fields, _m.labels...)
}

func (_m *ModuleFcPort) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleFcPort) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleFcPort) Run(logger *slog.Logger, col *Collector) {
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
			// Get WWNN, WWPN
			wwn := v.Get("wwn").String()
			var wwnn string
			var wwpn string
			if len(wwn) > 24 {
				wwnn = wwn[:23]
				wwpn = wwn[24:]
			}
			fcPortAttrs := metric.WithAttributes(
				attribute.String("fePort", v.Get("id").String()),
			)

			infoAttrs := metric.WithAttributes(
				attribute.String("slot.id", v.Get("slotNumber").String()),
				attribute.String("fePort.name", v.Get("name").String()),
				attribute.String("fc.wwnn", wwnn),
				attribute.String("fc.wwpn", wwpn),
			)
			for _, desc := range _m.desc {
				key := desc.Key
				switch key {
				case "info":
					observer.ObserveFloat64(observableMap[key], 1, clientAttrs, fcPortAttrs, infoAttrs)
				case "health.value":
					observer.ObserveFloat64(observableMap[key], v.Get(key).Float(), clientAttrs, fcPortAttrs)
				case "currentSpeed":
					observer.ObserveFloat64(observableMap[key], v.Get(key).Float(), clientAttrs, fcPortAttrs)
				}
			}
		}

		return nil
	}, observableArray...)

}
