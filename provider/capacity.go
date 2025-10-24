package provider

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils"

	"go.opentelemetry.io/otel/metric"
)

func init() {
	// Set Default
	moduleName := "capacity"
	SetDefaultProvider(moduleName, true)

	// Init Metrics Descriptions...
	var capacityMetricDescs = []*MetricDescriptor{
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

	// Init Option
	opt := api.NewUnityActionOptions("systemCapacity")
	for _, desc := range capacityMetricDescs {
		opt.Fields = append(opt.Fields, desc.Key)
	}

	registryProvider(moduleName, &capacityProvider{
		moduleName: moduleName,
		opts:       opt,
		desc:       capacityMetricDescs,
	})
}

type capacityProvider struct {
	moduleName string
	opts       *api.UnityActionOptions
	desc       []*MetricDescriptor
}

func (_pv *capacityProvider) Run(logger *slog.Logger, col *Collector) {
	meter := col.meterProvider.Meter(_pv.moduleName)
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
			logger.Debug("hostLabels not set")
			return nil
		}
		clientAttrs := metric.WithAttributes(append(col.resource.Attributes(), col.labels...)...)

		// Request Data
		data, err := client.GetInstances(_pv.opts)
		if err != nil {
			logger.Error("Failed to get", "error", err, "module", _pv.moduleName)
			return nil
		}

		// Capacity Attributes...
		for _, v := range data {
			for _, desc := range _pv.desc {
				key := desc.Key
				observer.ObserveFloat64(observableMap[key], utils.Bytes(v.Get(key).Int()).ToMiB(), clientAttrs)
			}
		}

		return nil
	}, observableArray...)

}
