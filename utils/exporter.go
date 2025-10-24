package utils

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

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
