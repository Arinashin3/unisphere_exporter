package provider

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unisphere_exporter/gounity"
	"unisphere_exporter/gounity/api"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	//"go.opentelemetry.io/otel/metric"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const serviceName = "unisphere_exporter"

var (
	ModuleOptions = make(map[string]Option)
	ProviderList  = make(map[string]Provider)
)

func SetDefaultProvider(name string, enabled bool) {
}

func FindDefaultProvider(key string) bool {
	for k, _ := range ModuleOptions {
		if k == key {
			return true
		}
	}
	return false
}

type Collector struct {
	ctx            context.Context
	resource       *resource.Resource
	labels         []attribute.KeyValue
	meterProvider  *sdkMetric.MeterProvider
	loggerProvider *sdkLog.LoggerProvider
	interval       time.Duration
	Client         *gounity.UnisphereClient
}

func NewCollector(ctx context.Context, attrs []attribute.KeyValue) (*Collector, error) {
	if res, err := resource.New(ctx, resource.WithAttributes(attrs...)); err != nil {
		return nil, err
	} else {
		return &Collector{resource: res}, nil
	}
}

func (_col *Collector) SetInterval(duration time.Duration) *Collector {
	_col.interval = duration
	return _col
}

func (_col *Collector) NewMeterProvider(interval time.Duration, exp *sdkMetric.Exporter) {
	_col.meterProvider = sdkMetric.NewMeterProvider(
		sdkMetric.WithResource(_col.resource),
		sdkMetric.WithReader(
			sdkMetric.NewPeriodicReader(*exp,
				sdkMetric.WithInterval(interval),
			),
		),
	)
}

func (_col *Collector) NewLoggerProvider(interval time.Duration, exp *sdkLog.Exporter) {
	_col.loggerProvider = sdkLog.NewLoggerProvider(
		sdkLog.WithResource(_col.resource),
		sdkLog.WithProcessor(
			sdkLog.NewBatchProcessor(*exp,
				sdkLog.WithExportInterval(interval),
			),
		),
	)
}

func (_col *Collector) Start(logger *slog.Logger) {
	// Init HostName
	opt := api.NewUnityActionOptions("system")
	opt.Fields = []string{"name"}
	data, err := _col.Client.GetInstances(opt)
	if err != nil {
		logger.Warn("cannot set labels", "error", err)
	} else {
		for _, v := range data {
			v.Get("name").String()
			_col.labels = append(_col.labels, attribute.String("host.name", v.Get("name").String()))
		}

	}

	for k, v := range ModuleOptions {
		if v != nil {
			continue
		}
		go ProviderList[k].Run(logger, _col)
	}
	select {}
}

func registryProvider(moduleName string, pv Provider) {
	ProviderList[moduleName] = pv
}

type Provider interface {
	Run(logger *slog.Logger, col *Collector)
}

// Back
type MetricDescriptor struct {
	Key      string
	Name     string
	Desc     string
	Unit     string
	TypeName string
}

func CreateMapMetricDescriptor(meter metric.Meter, mds []*MetricDescriptor, logger *slog.Logger) map[string]metric.Float64Observable {
	mdmap := make(map[string]metric.Float64Observable)
	var err error
	for _, md := range mds {
		var tmp metric.Float64Observable
		desc := metric.WithDescription(md.Desc)
		unit := metric.WithUnit(md.Unit)
		switch md.TypeName {
		case "counter":
			tmp, err = meter.Float64ObservableCounter(md.Name, desc, unit)
		case "gauge":
			tmp, err = meter.Float64ObservableGauge(md.Name, desc, unit)
		default:
			err = errors.New("unknown metric type")
		}
		if err != nil {
			logger.Warn("cannot create metric", "error", err, "metric_key", md.Key, "metric_type", md.TypeName)
		}
		mdmap[md.Key] = tmp
	}
	return mdmap

}
