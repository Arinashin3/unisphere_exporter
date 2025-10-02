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
	moduleName := "system"
	registProvider(moduleName, &systemProvider{moduleName: moduleName})
}

func (pv *systemProvider) IsDefaultEnabled() bool {
	return false
}

func (pv *systemProvider) NewProvider(cfg *config.UnisphereConfig, moduleName string, cl *ClientDesc) Provider {
	pvConf := cfg.Providers.System
	enabled := pvConf.GetEnabled(pv.IsDefaultEnabled())
	interval := pvConf.GetInterval()

	if !enabled {
		return nil
	}
	if MetricExporter == nil {
		return nil
	}
	mp := NewMeterProvider(serviceName, interval, MetricExporter)
	return &systemProvider{
		moduleName:    moduleName,
		interval:      interval,
		meterProvider: mp,
		clientDesc:    cl,
	}
}

type systemProvider struct {
	moduleName    string
	interval      time.Duration
	meterProvider *sdkMetric.MeterProvider
	clientDesc    *ClientDesc
}

var systemMetricDescs = []*MetricDescriptor{
	{
		Key:      "info",
		Name:     "unisphere_system_info",
		Desc:     "Information about unisphere system",
		Unit:     "",
		TypeName: "gauge",
	},
}

func (pv *systemProvider) Run(logger *slog.Logger) {
	logger.Info("Starting provider", "endpoint", pv.clientDesc.endpoint, "provider", pv.moduleName)
	meter := pv.meterProvider.Meter(pv.moduleName)
	uc := pv.clientDesc.client

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, systemMetricDescs, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// Client Attributes
		if pv.clientDesc.hostLabels == nil {
			return errors.New("hostLabels not set")
		}
		clientAttrs := metric.WithAttributes(pv.clientDesc.hostLabels...)

		// Request Data (MgmtInterface)
		mgmtData, err := uc.GetMgmtInterfaceInstances([]string{"ipAddress"}, nil)
		if err != nil {
			logger.Error("Failed to get system", "error", err)
			return nil
		}

		var ipaddr string
		for _, entry := range mgmtData.Entries {
			content := entry.Content
			ipaddr = content.IpAddress
		}
		// Request Data (BasicSystemInfo)
		data, err := uc.GetBasicSystemInfoInstances()
		if err != nil {
			logger.Error("Failed to get system", "error", err)
			return nil
		}

		// System Attributes...
		for _, entry := range data.Entries {
			content := entry.Content
			infoAttrs := metric.WithAttributes(attribute.String("product.name", content.Model), attribute.String("firmware.version", content.SoftwareFullVersion), attribute.String("ip.address", ipaddr))
			observer.ObserveFloat64(observableMap["info"], 1, clientAttrs, infoAttrs)
		}

		return nil
	}, observableArray...)

}
