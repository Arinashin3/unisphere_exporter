package provider

import (
	"context"
	"log/slog"
	"time"
	"unisphere_exporter/config"

	"go.opentelemetry.io/otel/log"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
)

func init() {
	moduleName := "event"
	registProvider(moduleName, &eventProvider{moduleName: moduleName})
}

func (pv *eventProvider) IsDefaultEnabled() bool {
	return true
}

func (pv *eventProvider) NewProvider(cfg *config.UnisphereConfig, moduleName string, cl *ClientDesc) Provider {
	pvConf := cfg.Providers.Event
	enabled := pvConf.GetEnabled(pv.IsDefaultEnabled())
	interval := pvConf.GetInterval()

	if !enabled {
		return nil
	}
	if LogExporter == nil {
		return nil
	}
	lp := NewLoggerProvider(serviceName, interval, LogExporter)
	return &eventProvider{
		moduleName:     moduleName,
		interval:       interval,
		level:          pvConf.Level,
		loggerProvider: lp,
		clientDesc:     cl,
	}
}

type eventProvider struct {
	moduleName     string
	interval       time.Duration
	level          int
	loggerProvider *sdkLog.LoggerProvider
	clientDesc     *ClientDesc
}

func (pv *eventProvider) Run(logger *slog.Logger) {
	logger.Info("Starting provider", "endpoint", pv.clientDesc.endpoint, "provider", pv.moduleName)
	ctx := context.Background()
	ctime := time.Now().Add(-time.Hour).UTC()
	uc := pv.clientDesc.client
	lp := pv.loggerProvider

	for {

		pvlogger := lp.Logger(pv.moduleName, log.WithInstrumentationAttributes(pv.clientDesc.hostLabels...))
		var fields = []string{
			"creationTime",
			"severity",
			"messageId",
			"message",
			"source",
		}
		filters := []string{
			"creationTime gt \"" + ctime.Format("2006-01-02T15:04:05.000Z") + "\"",
		}
		data, err := uc.GetEventInstances(fields, filters)
		if err != nil {
			logger.Error("Error to GET EventLog", "err", err)
			time.Sleep(pv.interval)
			continue
		}
		if data == nil {
			time.Sleep(pv.interval)
			continue
		}

		for _, entry := range data.Entries {
			record := log.Record{}
			content := entry.Content
			if pv.level > int(content.Severity) {
				continue
			}

			record.SetTimestamp(content.CreationTime)
			record.SetObservedTimestamp(content.CreationTime)
			record.SetBody(log.StringValue(content.Message))
			record.AddAttributes(
				log.String("level", content.Severity.String()),
				log.String("message.id", content.MessageId),
				log.String("source", content.Source),
			)
			pvlogger.Emit(ctx, record)

		}
		ctime = data.Updated.UTC()

		time.Sleep(pv.interval)
	}

}
