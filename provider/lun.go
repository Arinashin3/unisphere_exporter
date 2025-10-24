package provider

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils"

	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/metric"
)

type lunProvider struct {
	moduleName string
	opts       *api.UnityActionOptions
	desc       []*MetricDescriptor
}

func init() {
	// Set Default
	moduleName := "lun"
	SetDefaultProvider(moduleName, false)

	// Init Metrics Descriptions...
	var lunMetricDescs = []*MetricDescriptor{
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

	// Init Option
	opt := api.NewUnityActionOptions(moduleName)
	for _, desc := range lunMetricDescs {
		opt.Fields = append(opt.Fields, desc.Key)
	}
	opt.Fields = append(opt.Fields, "name", "wwn")

	registryProvider(moduleName, &lunProvider{
		moduleName: moduleName,
		opts:       opt,
		desc:       lunMetricDescs,
	})
}

func (_pv *lunProvider) Run(logger *slog.Logger, col *Collector) {
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
			lunAttrs := metric.WithAttributes(attribute.String("lun.name", v.Get("name").String()), attribute.String("lun.wwn", v.Get("wwn").String()))
			for _, desc := range _pv.desc {
				key := desc.Key
				observer.ObserveFloat64(observableMap[key], utils.Bytes(v.Get(key).Int()).ToMiB(), clientAttrs, lunAttrs)
			}
		}

		return nil
	}, observableArray...)

}
