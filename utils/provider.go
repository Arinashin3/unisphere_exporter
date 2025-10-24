package utils

import (
	"time"

	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProvider(interval time.Duration, exp *sdkMetric.Exporter, res *resource.Resource) *sdkMetric.MeterProvider {
	return sdkMetric.NewMeterProvider(
		sdkMetric.WithResource(res),
		sdkMetric.WithReader(
			sdkMetric.NewPeriodicReader(*exp,
				sdkMetric.WithInterval(interval),
			),
		),
	)
}
func NewLoggerProvider(interval time.Duration, exp *sdkLog.Exporter, res *resource.Resource) *sdkLog.LoggerProvider {
	return sdkLog.NewLoggerProvider(
		sdkLog.WithResource(res),
		sdkLog.WithProcessor(
			sdkLog.NewBatchProcessor(*exp,
				sdkLog.WithExportInterval(interval),
			),
		),
	)
}
