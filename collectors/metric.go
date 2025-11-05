package collectors

import (
	"context"
	"log/slog"
	"strings"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils"

	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gopkg.in/yaml.v3"
)

type ModuleMetric struct {
	// Module's Information
	name     string
	defaults bool
	desc     []*MetricDescriptor

	// Configuration File
	Enabled *bool    `yaml:"enabled"`
	Paths   []string `yaml:"paths"`
}

func init() {
	key := "metric"
	RegisterModule(key, NewMetric())
}

func NewMetric() *ModuleMetric {
	return &ModuleMetric{
		defaults: false,
	}
}

func (_m *ModuleMetric) Init(key string) {
	_m.name = key
}

func (_m *ModuleMetric) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleMetric) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		if len(_m.Paths) > 0 {
			_m.Enabled = new(bool)
			*_m.Enabled = true
		} else {
			_m.Enabled = &_m.defaults
		}
	}
	// TODO
	// Logger
	if len(_m.Paths) == 0 {
	}
}

func (_m *ModuleMetric) Run(logger *slog.Logger, col *Collector) {
	meter := col.meterProvider.Meter(_m.name)
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
		for _, path := range _m.Paths {
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
					logger.Info("SKIP THIS METRIC: this metric is not output number", "module", _m.name, "path", path)
					continue
				case 7:
					mType = "counter"
				case 8:
					mType = "counter"
				}
				tmp := "unisphere_" + strings.Replace(strings.ToLower(v.Get("path").String()), ".*.", "_", -1)

				metricPaths = append(metricPaths, v.Get("path").String())
				_m.desc = append(_m.desc, &MetricDescriptor{
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
	observableMap = CreateMapMetricDescriptor(meter, _m.desc, logger)
	//
	//// Register Metrics for Observables...
	var observableArray []metric.Observable
	for _, obserable := range observableMap {
		observableArray = append(observableArray, obserable)
	}

	// Metric Realtime Query Maximum Paths == 48
	if len(observableArray) > 48 {
		logger.Error("Too Many Paths", "provider", _m.name, "path_count", len(metricPaths))
		return
	}
	createQidOpts := api.NewUnityActionOptions(string(api.UnityMetricRealTimeQuery))
	logger.Info("Create Metric Query", "provider", _m.name, "path_count", len(metricPaths))

	// Get Query ID
	var qid string
	if qid, err = client.PostMetricRealTimeQuery(createQidOpts, _m.Paths, col.interval); err != nil {
		logger.Warn("cannot create metric", "err", err)
		return
	} else if qid == "" {
		logger.Warn("cannot get query id", "err", "qid is empty")
	}

	//// Callback
	meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		if qid == "" {
			logger.Info("Recreate the Metric Realtime Query", "provider", _m.name, "path_count", len(metricPaths))
			if qid, err = client.PostMetricRealTimeQuery(createQidOpts, _m.Paths, col.interval); err != nil {
				logger.Warn("cannot create metric", "err", err)
				return nil
			}
		}

		// Set Attributes
		//if col.labels == nil {
		//	logger.Debug("hostLabels not set")
		//	return nil
		//}
		clientAttrs := metric.WithAttributes(append(col.resource.Attributes(), col.labels...)...)

		// Request Data
		opts := api.NewUnityActionOptions("metricQueryResult")
		opts.Filters = []string{"queryId eq " + qid}
		var data []gjson.Result
		data, err = client.GetInstances(opts)
		if err != nil {
			logger.Error("Failed to get metric", "error", err)
			col.success = false
			qid = ""
			return nil
		}
		col.success = true

		// Parsing Metric &
		for _, content := range data {
			// Create Label Name
			var key = content.Get("path").String()
			var labelKeys []string
			var preString string
			for _, v := range strings.Split(key, ".") {
				if v == "*" {
					labelKeys = append(labelKeys, preString)
				}
				preString = v
			}

			// Get Values...
			result := utils.ParseMetric(content.Get("values"))
			for _, r := range result {
				var metricLabels []attribute.KeyValue
				for i, lname := range labelKeys {
					metricLabels = append(metricLabels, attribute.String(lname, r.Labels[i]))
				}
				observer.ObserveFloat64(observableMap[key], r.Value.Float(), clientAttrs, metric.WithAttributes(metricLabels...))
			}
		}

		return nil
	}, observableArray...)
}
