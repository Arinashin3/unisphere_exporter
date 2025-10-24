package provider

import (
	"context"
	"log/slog"
	"unisphere_exporter/gounity/api"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func init() {
	moduleName := "basicSystemInfo"

	SetDefaultProvider(moduleName, true)

	// Init Metrics Descriptions...
	var basicSystemInfoMetricDescs = []*MetricDescriptor{
		{
			Key:      "info",
			Name:     "unisphere_basic_system_info",
			Desc:     "Information about unisphere basicSystemInfo",
			Unit:     "",
			TypeName: "gauge",
		},
	}

	// Init Option
	opt := api.NewUnityActionOptions("basicSystemInfo")
	opt.Fields = []string{"model", "softwareFullVersion"}

	registryProvider(moduleName, &basicSystemInfoProvider{
		moduleName: moduleName,
		opts:       opt,
		desc:       basicSystemInfoMetricDescs,
	})
}

type basicSystemInfoProvider struct {
	moduleName string
	opts       *api.UnityActionOptions
	desc       []*MetricDescriptor
}

func (_pv *basicSystemInfoProvider) Run(logger *slog.Logger, col *Collector) {
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

		// System Attributes...
		for _, v := range data {
			infoAttrs := metric.WithAttributes(attribute.String("product.name", v.Get("model").String()), attribute.String("firmware.version", v.Get("softwareFullVersion").String()))
			observer.ObserveFloat64(observableMap["info"], 1, clientAttrs, infoAttrs)
		}

		return nil
	}, observableArray...)

}
