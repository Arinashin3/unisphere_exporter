package provider

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unisphere_exporter/config"

	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	moduleName := "lun"
	registProvider(moduleName, &lunProvider{moduleName: moduleName})
}

func (pv *lunProvider) IsDefaultEnabled() bool {
	return false
}

func (pv *lunProvider) NewProvider(cfg *config.UnisphereConfig, moduleName string, cl *ClientDesc) Provider {
	pvConf := cfg.Providers.Lun
	enabled := pvConf.GetEnabled(pv.IsDefaultEnabled())
	interval := pvConf.GetInterval()

	if !enabled {
		return nil
	}
	if MetricExporter == nil {
		return nil
	}
	mp := NewMeterProvider(serviceName, interval, MetricExporter)
	return &lunProvider{
		moduleName:    moduleName,
		interval:      interval,
		meterProvider: mp,
		clientDesc:    cl,
	}
}

type lunProvider struct {
	moduleName    string
	interval      time.Duration
	meterProvider *sdkMetric.MeterProvider
	clientDesc    *ClientDesc
}

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

func (pv *lunProvider) Run(logger *slog.Logger) {
	logger.Info("Starting provider", "endpoint", pv.clientDesc.endpoint, "provider", pv.moduleName)
	meter := pv.meterProvider.Meter(pv.moduleName)
	uc := pv.clientDesc.client

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, lunMetricDescs, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Request Fields
	var paramsFields []string
	for _, v := range lunMetricDescs {
		paramsFields = append(paramsFields, v.Key)
	}
	paramsFields = append(paramsFields, "name", "wwn")

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// Client Attributes
		if pv.clientDesc.hostLabels == nil {
			return errors.New("hostLabels not set")
		}
		clientAttrs := metric.WithAttributes(pv.clientDesc.hostLabels...)

		// Request Data
		data, err := uc.GetLunInstances(paramsFields, nil)
		if err != nil {
			logger.Error("Failed to get lun", "error", err)
			return nil
		}

		// Lun Attributes...
		for _, entry := range data.Entries {
			content := entry.Content
			lunAttrs := metric.WithAttributes(attribute.String("lun.name", content.Name), attribute.String("lun.wwn", content.Wwn))
			observer.ObserveFloat64(observableMap["sizeTotal"], content.SizeTotal.ToMiB(), clientAttrs, lunAttrs)
			observer.ObserveFloat64(observableMap["sizeUsed"], content.SizeUsed.ToMiB(), clientAttrs, lunAttrs)
			observer.ObserveFloat64(observableMap["sizeAllocated"], content.SizeAllocated.ToMiB(), clientAttrs, lunAttrs)
			observer.ObserveFloat64(observableMap["sizePreallocated"], content.SizePreallocated.ToMiB(), clientAttrs, lunAttrs)
		}

		return nil
	}, observableArray...)

}
