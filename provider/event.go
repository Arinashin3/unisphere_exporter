package provider

import (
	"encoding/json"
	"log/slog"
	"time"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils/enum"

	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/log"
)

func init() {
	moduleName := "event"
	SetDefaultProvider(moduleName, true)
	opt := api.NewUnityActionOptions(moduleName)
	startTime := time.Now().Add(-12 * time.Hour).UTC()
	opt.Fields = []string{
		"creationTime",
		"severity",
		"messageId",
		"message",
		"source",
	}
	opt.Filters = []string{
		"creationTime gt \"" + startTime.Format("2006-01-02T15:04:05.000Z") + "\"",
	}
	registryProvider(moduleName, &eventProvider{
		moduleName: moduleName,
		opt:        opt,
	})
}

type eventProvider struct {
	moduleName string
	opt        *api.UnityActionOptions
	level      int
}

func (_pv *eventProvider) Run(logger *slog.Logger, col *Collector) {
	opt := *_pv.opt
	ctime := time.Now().Add(-1 * time.Hour).UTC()
	client := col.Client
	lp := col.loggerProvider

	for {
		pvlogger := lp.Logger(_pv.moduleName, log.WithInstrumentationAttributes(col.labels...))
		opt.Filters = []string{
			"creationTime gt \"" + ctime.Format("2006-01-02T15:04:05.000Z") + "\"",
		}

		tmpTime := time.Now().UTC()
		data, err := client.GetInstances(&opt)
		if err != nil {
			logger.Error("Error to GET EventLog", "err", err)
			time.Sleep(col.interval)
			continue
		}
		if data == nil {
			time.Sleep(col.interval)
			continue
		}

		for _, v := range data {
			record := log.Record{}
			if _pv.level > int(v.Get("severity").Int()) {
				continue
			}

			record.SetTimestamp(v.Get("creationTime").Time())
			logBody := struct {
				Source    string `json:"source"`
				Message   string `json:"message"`
				MessageId string `json:"message_id"`
			}{
				v.Get("source").String(),
				v.Get("message").String(),
				v.Get("messageId").String(),
			}
			jsonBody, _ := json.Marshal(logBody)
			body := gjson.ParseBytes(jsonBody).String()
			record.SetBody(log.StringValue(body))
			record.AddAttributes(
				log.String("level", enum.SeverityEnum(v.Get("severity").Int()).String()),
			)
			pvlogger.Emit(col.ctx, record)

		}

		ctime = tmpTime
		time.Sleep(col.interval)
	}

}
