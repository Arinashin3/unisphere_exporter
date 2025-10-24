package provider

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unisphere_exporter/config"

	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	moduleName := "capacity"
	registProvider(moduleName, &capacityProvider{moduleName: moduleName})
}

func (pv *capacityProvider) IsDefaultEnabled() bool {
	return false
}

func (pv *capacityProvider) NewProvider(cfg *config.UnisphereConfig, moduleName string, cl *ClientDesc) Provider {
	pvConf := cfg.Providers.Capacity
	enabled := pvConf.GetEnabled(pv.IsDefaultEnabled())
	interval := pvConf.GetInterval()

	if !enabled {
		return nil
	}
	if MetricExporter == nil {
		return nil
	}
	mp := NewMeterProvider(serviceName, interval, MetricExporter)
	return &capacityProvider{
		moduleName:    moduleName,
		interval:      interval,
		meterProvider: mp,
		clientDesc:    cl,
	}
}

type capacityProvider struct {
	moduleName    string
	interval      time.Duration
	meterProvider *sdkMetric.MeterProvider
	clientDesc    *ClientDesc
}

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

func (pv *capacityProvider) Run(logger *slog.Logger) {
	logger.Info("Starting provider", "endpoint", pv.clientDesc.endpoint, "provider", pv.moduleName)
	meter := pv.meterProvider.Meter(pv.moduleName)
	uc := pv.clientDesc.client

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, capacityMetricDescs, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Request Fields
	var paramsFields = []string{"sizeTotal", "sizeUsed", "sizeFree", "sizePreallocated", "totalLogicalSize"}

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// Client Attributes
		if pv.clientDesc.hostLabels == nil {
			return errors.New("hostLabels not set")
		}
		clientAttrs := metric.WithAttributes(pv.clientDesc.hostLabels...)

		// Request Data
		data, err := uc.GetSystemCapacityInstances(paramsFields, nil)
		if err != nil {
			logger.Error("Failed to get capacity", "error", err)
			return nil
		}

		// Capacity Attributes...
		for _, entry := range data.Entries {
			content := entry.Content
			observer.ObserveFloat64(observableMap["sizeTotal"], content.SizeTotal.ToMiB(), clientAttrs)
			observer.ObserveFloat64(observableMap["sizeUsed"], content.SizeUsed.ToMiB(), clientAttrs)
			observer.ObserveFloat64(observableMap["sizeFree"], content.SizeFree.ToMiB(), clientAttrs)
			observer.ObserveFloat64(observableMap["sizePreallocated"], content.SizePreallocated.ToMiB(), clientAttrs)
			observer.ObserveFloat64(observableMap["totalLogicalSize"], content.TotalLogicalSize.ToMiB(), clientAttrs)
		}

		return nil
	}, observableArray...)

}
