package provider

import (
	"context"
	"log/slog"
	"strings"
	"unisphere_exporter/gounity/api"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func init() {
	moduleName := "metric"
	SetDefaultProvider(moduleName, true)

	registryProvider(moduleName, &metricProvider{
		moduleName: moduleName,
		paths:      []string{},
	})
}

func (pv)

type ModuleOptionMetric struct {
	Enabled bool
	Paths   []string
}

func NewOptionMetric() *ModuleOptionMetric {
	return &ModuleOptionMetric{
		Enabled: false,
		Paths:   []string{},
	}
}

type metricProvider struct {
	moduleName string
	paths      []string
	desc       []*MetricDescriptor
}

func (_pv *metricProvider) Run(logger *slog.Logger, col *Collector) {
	meter := col.meterProvider.Meter(_pv.moduleName)
	client := col.Client

	// Get Metric List...
	descOpts := api.NewUnityActionOptions("metric")
	descOpts.Fields = []string{"name", "path", "type", "unitDisplayString", "description"}
	descOpts.Filters = []string{
		"isRealtimeAvailable eq true",
	}
	descData, err := client.GetInstances(descOpts)
	if err != nil {
		logger.Warn("cannot get metric values", "err", err)
		return
	}

	// Create Metric Descriptions...
	var metricPaths []string
	for _, v := range descData {
		for _, path := range _pv.paths {
			var match bool
			// When last char is '%', remove it and find contain from metrics
			// others are find match.
			if string(path[len(path)-1]) == "%" {
				pattern := strings.Replace(path, "%", "", -1)
				match = strings.Contains(v.Get("path").String(), pattern)
			} else {
				if path == v.Get("path").String() {
					match = true
				}
			}
			if match {
				var mType string
				switch v.Get("type").Int() {
				case 2:
					mType = "counter"
				case 3:
					mType = "counter"
				case 4:
					mType = "gauge"
				case 5:
					mType = "gauge"
				case 6:
					logger.Info("SKIP THIS METRIC: this metric is not output number", "module", _pv.moduleName, "path", path)
					continue
				case 7:
					mType = "counter"
				case 8:
					mType = "counter"
				}
				tmp := "unisphere_" + strings.Replace(strings.ToLower(v.Get("path").String()), ".*.", "_", -1)

				metricPaths = append(metricPaths, v.Get("path").String())
				_pv.desc = append(_pv.desc, &MetricDescriptor{
					Key:      v.Get("path").String(),
					Name:     strings.Replace(tmp, ".", "_", -1),
					Desc:     v.Get("description").String(),
					Unit:     strings.ToLower(v.Get("unitDisplayString").String()),
					TypeName: mType,
				})
			}
		}
	}

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
		data, err := client.GetInstances(descOpts)
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
