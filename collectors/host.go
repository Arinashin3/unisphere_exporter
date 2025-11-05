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
	key := "host"
	RegisterModule(key, NewHost())
}

type ModuleHost struct {
	// Module's Information
	name     string
	opts     *api.UnityActionOptions
	desc     []*MetricDescriptor
	defaults bool
	labels   []string

	// Configuration File
	Enabled *bool `yaml:"enabled"`
}

func NewHost() *ModuleHost {
	return &ModuleHost{
		defaults: false,
	}
}

func (_m *ModuleHost) Init(key string) {
	_m.name = key
	_m.desc = []*MetricDescriptor{
		{
			Key:      "info",
			Name:     "unisphere_host_info",
			Desc:     "information of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		// Host's Metrics...
		{
			Key:      "health.value",
			Name:     "unisphere_host_health",
			Desc:     "Health of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		// Initiator's Metrics...
		{
			Key:      "initiator.health",
			Name:     "unisphere_host_initiator_health",
			Desc:     "Health of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		{
			Key:      "initiator.path",
			Name:     "unisphere_host_initiator_path",
			Desc:     "information of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
		// Lun's Metrics...
		// This Metric means mapping with luns...
		{
			Key:      "host.lun.map",
			Name:     "unisphere_host_lun_map",
			Desc:     "information of the associated resource",
			Unit:     "",
			TypeName: "gauge",
		},
	}

	_m.opts = api.NewUnityActionOptions("host")

	// Output Fields...
	_m.opts.Fields = []string{
		// Host(target)'s Information
		"id",
		"name",
		"osType",
		"health.value",
		// Fibre Channel Initiator
		"fcHostInitiators.health.value",
		"fcHostInitiators.initiatorId",
		"fcHostInitiators.paths.fcPort.id",
		// iSCSI Initiator
		"iscsiHostInitiators.health.value",
		"iscsiHostInitiators.initiatorId",
		"iscsiHostInitiators.paths.iscsiPortal.ethernetPort.id",
		// Lun
		"hostLUNs.lun.id",
	}
}

func (_m *ModuleHost) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleHost) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleHost) Run(logger *slog.Logger, col *Collector) {
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

		// Parse Data
		for _, v := range data {

			targetAttrs := metric.WithAttributes(
				attribute.String("target.id", v.Get("id").String()),
			)

			// For Info Observer
			infoAttrs := metric.WithAttributes(
				attribute.String("target.name", v.Get("name").String()),
				attribute.String("target.os", v.Get("osType").String()),
			)
			observer.ObserveFloat64(observableMap["info"], 1, clientAttrs, targetAttrs, infoAttrs)
			observer.ObserveFloat64(observableMap["health.value"], v.Get("health.value").Float(), clientAttrs, targetAttrs)

			// Fibre Channel Initiators...
			if v.Get("fcHostInitiators").Exists() {
				for _, initiator := range v.Get("fcHostInitiators").Array() {
					initiatorId := initiator.Get("initiatorId").String()
					wwn := attribute.String("fc.wwn", initiatorId)
					observer.ObserveFloat64(observableMap["initiator.health"], initiator.Get("health.value").Float(), clientAttrs, targetAttrs, metric.WithAttributes(wwn))
					for _, p := range initiator.Get("paths").Array() {
						portId := attribute.String("fePort", p.Get("fcPort.id").String())
						observer.ObserveFloat64(observableMap["initiator.path"], 1, clientAttrs, targetAttrs, metric.WithAttributes(wwn, portId))
					}
				}
			}
			// iSCSI Initiators...
			if v.Get("iscsiHostInitiators").Exists() {
				for _, initiator := range v.Get("iscsiHostInitiators").Array() {
					initiatorId := initiator.Get("initiatorId").String()
					iqn := attribute.String("iscsi.iqn", initiatorId)
					observer.ObserveFloat64(observableMap["initiator.health"], initiator.Get("health.value").Float(), clientAttrs, targetAttrs, metric.WithAttributes(iqn))
					for _, p := range initiator.Get("paths").Array() {
						portId := attribute.String("fePort", p.Get("iscsiPortal.ethernetPort.id").String())
						observer.ObserveFloat64(observableMap["initiator.path"], 1, clientAttrs, targetAttrs, metric.WithAttributes(iqn, portId))
					}
				}
			}
			// Lun Mapping...
			if v.Get("hostLUNs").Exists() {
				for _, lun := range v.Get("hostLUNs").Array() {
					lunAttrs := metric.WithAttributes(attribute.String("lun.id", lun.Get("lun.id").String()))
					observer.ObserveFloat64(observableMap["host.lun.map"], 1, clientAttrs, targetAttrs, lunAttrs)
				}
			}
		}

		return nil
	}, observableArray...)

}
