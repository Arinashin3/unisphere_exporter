package provider

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unisphere_exporter/config"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	registProvider("metric_a", &metricProvider{moduleName: "metric_a"})
	registProvider("metric_b", &metricProvider{moduleName: "metric_b"})
	registProvider("metric_c", &metricProvider{moduleName: "metric_c"})
}

func (pv *metricProvider) IsDefaultEnabled() bool {
	return false
}

func (pv *metricProvider) NewProvider(cfg *config.UnisphereConfig, moduleName string, cl *ClientDesc) Provider {
	var pvConf *config.ProviderMetric
	switch moduleName {
	case "metric_a":
		pvConf = cfg.Providers.Metric_A
	case "metric_b":
		pvConf = cfg.Providers.Metric_B
	case "metric_c":
		pvConf = cfg.Providers.Metric_C
	}
	if pvConf == nil {
		return nil
	}
	if pvConf.Paths == nil {
		return nil
	}

	enabled := pvConf.GetEnabled(pv.IsDefaultEnabled())
	interval := pvConf.GetInterval()

	if !enabled {
		return nil
	}
	if MetricExporter == nil {
		return nil
	}
	mp := NewMeterProvider(serviceName, interval, MetricExporter)
	return &metricProvider{
		moduleName:    moduleName,
		queryId:       0,
		interval:      interval,
		paths:         pvConf.Paths,
		meterProvider: mp,
		clientDesc:    cl,
	}
}

type metricProvider struct {
	moduleName    string
	interval      time.Duration
	queryId       int
	paths         []string
	meterProvider *sdkMetric.MeterProvider
	clientDesc    *ClientDesc
}

func (pv *metricProvider) Run(logger *slog.Logger) {
	logger.Info("Starting provider", "endpoint", pv.clientDesc.endpoint, "provider", pv.moduleName)
	meter := pv.meterProvider.Meter(pv.moduleName)
	uc := pv.clientDesc.client

	// Get Metric Descriptions from Unisphere API...
	filters := []string{
		"isRealtimeAvailable eq true",
	}
	metricData, err := uc.GetMetricInstances([]string{"name", "path", "type", "unitDisplayString", "description"}, filters)
	if err != nil {
		logger.Error("Failed to get metric instances", "provider", pv.moduleName, "error", err)
		return
	}

	paths := pv.paths
	var metricDescList []*MetricDescriptor
	var metricPaths []string

	for _, entry := range metricData.Entries {
		content := entry.Content
		for _, path := range paths {
			var match bool
			if string(path[len(path)-1]) == "%" {
				pattern := strings.Replace(path, "%", "", -1)
				match = strings.Contains(content.Path, pattern)
			} else {
				if path == content.Path {
					match = true
				}
			}
			if match {
				var mType string
				switch content.Type {
				case 2:
					mType = "counter"
				case 3:
					mType = "counter"
				case 4:
					mType = "gauge"
				case 5:
					mType = "gauge"
				case 6:
					logger.Info("SKIP METRIC: this metric's value is not number", "provider", pv.moduleName, "path", content.Path)
					continue
				case 7:
					mType = "counter"
				case 8:
					mType = "counter"
				}
				tmp := "unisphere_" + strings.Replace(strings.ToLower(content.Path), ".*.", "_", -1)

				metricPaths = append(metricPaths, content.Path)
				metricDescList = append(metricDescList, &MetricDescriptor{
					Key:      content.Path,
					Name:     strings.Replace(tmp, ".", "_", -1),
					Desc:     content.Description,
					Unit:     strings.ToLower(content.UnitDisplayString),
					TypeName: mType,
				})
			}
		}
	}

	// Register Metrics...
	var observableMap map[string]metric.Float64Observable
	observableMap = CreateMapMetricDescriptor(meter, metricDescList, logger)

	// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Metric Realtime Query Maximum Paths == 48
	if len(observableArray) > 48 {
		logger.Error("Too Many Paths", "provider", pv.moduleName, "path_count", len(metricPaths))
		return
	}
	logger.Info("Create Metric Query", "provider", pv.moduleName, "path_count", len(metricPaths))

	realtimeQuery, err := uc.PostMetricRealTimeQueryInstances(metricPaths, pv.interval)
	if err != nil {
		logger.Error("Failed to post metric query", "provider", pv.moduleName, "error", err)
	}
	pv.queryId = realtimeQuery.Content.Id

	// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		if pv.queryId == 0 {
			logger.Info("Recreate the Metric Realtime Query", "provider", pv.moduleName, "path_count", len(metricPaths))
			realtimeQuery, err = uc.PostMetricRealTimeQueryInstances(metricPaths, pv.interval)
			if err != nil {
				logger.Error("Failed to post metric query", "provider", pv.moduleName, "error", err)
				return nil
			}
			pv.queryId = realtimeQuery.Content.Id
		}

		// Client Attributes
		if pv.clientDesc.hostLabels == nil {
			return errors.New("hostLabels not set")
		}
		clientAttrs := metric.WithAttributes(pv.clientDesc.hostLabels...)

		// Request Data
		data, err := uc.GetMetricQueryResultInstances(pv.queryId)
		if err != nil {
			logger.Error("Failed to get metric", "error", err)
			return nil
		}

		// Metric Attributes...
		for _, entry := range data.Entries {
			content := entry.Content

			// Create Label Name
			var labels []string
			var preString string
			for _, v := range strings.Split(content.Path, ".") {
				if v == "*" {
					labels = append(labels, preString)
				}
				preString = v
			}

			// Get Metric
			for k1, v1 := range content.Values.(map[string]interface{}) {
				if reflect.TypeOf(v1).Kind().String() != "map" {
					var f float64
					f, err = strconv.ParseFloat(v1.(string), 64)
					if err != nil {
						logger.Error("Failed to parse metric value", "provider", pv.moduleName, "path", content.Path, "error", err)
						continue
					}
					observer.ObserveFloat64(observableMap[content.Path], f, clientAttrs, metric.WithAttributes(attribute.String(labels[0], k1)))
					continue
				}
				for k2, v2 := range v1.(map[string]interface{}) {
					if reflect.TypeOf(v2).Kind().String() != "map" {
						var f float64
						f, err = strconv.ParseFloat(v2.(string), 64)
						if err != nil {
							logger.Error("Failed to parse metric value", "provider", pv.moduleName, "path", content.Path, "error", err)
							continue
						}
						observer.ObserveFloat64(observableMap[content.Path], f, clientAttrs, metric.WithAttributes(attribute.String(labels[0], k1), attribute.String(labels[1], k2)))
						continue
					}
					for k3, v3 := range v2.(map[string]interface{}) {
						if reflect.TypeOf(v3).Kind().String() != "map" {
							var f float64
							f, err = strconv.ParseFloat(v3.(string), 64)
							if err != nil {
								logger.Error("Failed to parse metric value", "provider", pv.moduleName, "path", content.Path, "error", err)
								continue
							}
							observer.ObserveFloat64(observableMap[content.Path], f, clientAttrs, metric.WithAttributes(attribute.String(labels[0], k1), attribute.String(labels[1], k2), attribute.String(labels[2], k3)))
							continue
						}
					}
				}

			}
		}

		return nil
	}, observableArray...)

}
