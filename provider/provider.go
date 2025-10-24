package provider

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"
	"unisphere_exporter/config"

	"github.com/Arinashin3/gounity"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const serviceName = "unisphere_exporter"

var (
	UsableProviders = make(map[string]Provider)
	Providers       []Provider
	MetricExporter  *sdkMetric.Exporter
	LogExporter     *sdkLog.Exporter
)

func NewMeterProvider(svName string, interval time.Duration, exp *sdkMetric.Exporter) *sdkMetric.MeterProvider {
	return sdkMetric.NewMeterProvider(
		sdkMetric.WithResource(resource.NewSchemaless(attribute.String("service.name", svName))),
		sdkMetric.WithReader(
			sdkMetric.NewPeriodicReader(*exp,
				sdkMetric.WithInterval(interval),
			),
		),
	)
}

func NewMetricExporter(ctx context.Context, mode string, endpoint string, insecure bool) (*sdkMetric.Exporter, error) {
	var exp sdkMetric.Exporter
	var err error
	switch mode {
	case "http":
		if insecure {
			exp, err = otlpmetrichttp.New(ctx,
				otlpmetrichttp.WithEndpointURL(endpoint),
				otlpmetrichttp.WithInsecure(),
			)
		} else {
			exp, err = otlpmetrichttp.New(ctx,
				otlpmetrichttp.WithEndpointURL(endpoint),
			)
		}
	case "grpc":
		if insecure {
			exp, err = otlpmetricgrpc.New(ctx,
				otlpmetricgrpc.WithEndpointURL(endpoint),
				otlpmetricgrpc.WithInsecure(),
			)
		} else {
			exp, err = otlpmetricgrpc.New(ctx,
				otlpmetricgrpc.WithEndpointURL(endpoint),
			)
		}
	}
	return &exp, err
}

func NewLoggerProvider(svName string, interval time.Duration, exp *sdkLog.Exporter) *sdkLog.LoggerProvider {
	return sdkLog.NewLoggerProvider(
		sdkLog.WithResource(resource.NewSchemaless(attribute.String("service.name", svName))),
		sdkLog.WithProcessor(
			sdkLog.NewBatchProcessor(*exp,
				sdkLog.WithExportInterval(interval),
			),
		),
	)
}

func NewLogExporter(ctx context.Context, mode string, endpoint string, insecure bool) (*sdkLog.Exporter, error) {
	var exp sdkLog.Exporter
	var err error
	switch mode {
	case "http":
		if insecure {
			exp, err = otlploghttp.New(ctx,
				otlploghttp.WithEndpointURL(endpoint),
				otlploghttp.WithInsecure(),
			)
		} else {
			exp, err = otlploghttp.New(ctx,
				otlploghttp.WithEndpointURL(endpoint),
			)
		}
	case "grpc":
		if insecure {
			exp, err = otlploggrpc.New(ctx,
				otlploggrpc.WithEndpointURL(endpoint),
				otlploggrpc.WithInsecure(),
			)
		} else {
			exp, err = otlploggrpc.New(ctx,
				otlploggrpc.WithEndpointURL(endpoint),
			)
		}
	}
	return &exp, err
}

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

func RegistryProviders(cfg *config.UnisphereConfig, logger *slog.Logger) bool {
	var err error
	var success = true

	// Define MetricExporter
	ctx := context.Background()
	expEndpoint := cfg.Server.Metrics.Endpoint + cfg.Server.Metrics.Api_Path
	expMode := cfg.Server.Metrics.Mode
	expInsecure, _ := strconv.ParseBool(cfg.Server.Metrics.Insecure)
	if expEndpoint != "" {
		MetricExporter, err = NewMetricExporter(ctx, expMode, expEndpoint, expInsecure)
		if err != nil {
			logger.Error("Failed to create metric exporter.", "error", err)
			success = false
		}

	}

	// Define LogExporter
	expEndpoint = cfg.Server.Logs.Endpoint + cfg.Server.Logs.Api_Path
	expMode = cfg.Server.Logs.Mode
	expInsecure, _ = strconv.ParseBool(cfg.Server.Logs.Insecure)
	if expEndpoint != "" {
		LogExporter, err = NewLogExporter(ctx, expMode, expEndpoint, expInsecure)
		if err != nil {
			logger.Error("Failed to create the Log Exporter...", "error", err)
			success = false
		}
	}

	for _, clientConf := range cfg.Clients {
		var endpoint string
		var customLabels []attribute.KeyValue
		var insecure bool

		endpoint = clientConf.Endpoint
		for k, v := range clientConf.Labels {
			customLabels = append(customLabels, attribute.String(k, v))
		}
		username, password := cfg.SearchAuth(clientConf.Auth)
		if username == "" || password == "" {
			logger.Error("Cannot found the authentication credentials.", "auth", clientConf.Auth)
			success = false
		}
		insecure, _ = strconv.ParseBool(clientConf.Insecure)
		cm := gounity.NewClient(endpoint, username, password, insecure)

		cl := &ClientDesc{
			endpoint:     endpoint,
			customLabels: customLabels,
			hostLabels:   nil,
			client:       cm,
		}
		_ = UpdateAttributes(cl)

		for k, pv := range UsableProviders {
			tmp := pv.NewProvider(cfg, k, cl)
			if tmp != nil {
				Providers = append(Providers, tmp)
			}
		}
	}
	return success
}

func RunProviders(logger *slog.Logger) {
	for _, pv := range Providers {
		go pv.Run(logger)
	}
	select {}
}

func UpdateAttributes(cl *ClientDesc) error {
	data, err := cl.client.GetSystemInstances([]string{"name", "serialNumber"}, nil)
	if data == nil {
		return errors.New("cannot to Update Attributes")
	}
	if err != nil {
		return err
	}

	var tmp []attribute.KeyValue
	if cl.customLabels != nil {
		tmp = cl.customLabels
	}
	for _, entry := range data.Entries {
		content := entry.Content
		tmp = append(tmp, attribute.String("host.name", content.Name))
		tmp = append(tmp, attribute.String("instance", content.SerialNumber))
	}
	cl.hostLabels = tmp

	return nil
}

type ClientDesc struct {
	endpoint     string
	customLabels []attribute.KeyValue
	hostLabels   []attribute.KeyValue
	client       *gounity.UnisphereClient
}

type Provider interface {
	NewProvider(cfg *config.UnisphereConfig, moduleName string, desc *ClientDesc) Provider
	Run(logger *slog.Logger)
}

func registProvider(moduleName string, pv Provider) error {
	UsableProviders[moduleName] = pv
	return nil
}
