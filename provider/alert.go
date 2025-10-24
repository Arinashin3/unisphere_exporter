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
	moduleName := "alert"
	SetDefaultProvider(moduleName, true)
	opt := api.NewUnityActionOptions(moduleName)
	startTime := time.Now().Add(-1 * time.Hour).UTC()
	opt.Fields = []string{
		"timestamp",
		"severity",
		"messageId",
		"message",
	}
	opt.Filters = []string{
		"timestamp gt \"" + startTime.Format("2006-01-02T15:04:05.000Z") + "\"",
	}
	registryProvider(moduleName, &alertProvider{
		moduleName: moduleName,
		opt:        opt,
	})

}

type alertProvider struct {
	moduleName string
	opt        *api.UnityActionOptions
	level      int
}

func (_pv *alertProvider) Run(logger *slog.Logger, col *Collector) {
	opt := *_pv.opt
	ctime := time.Now().Add(-1 * time.Hour).UTC()
	client := col.Client
	lp := col.loggerProvider

	for {

		pvlogger := lp.Logger(_pv.moduleName, log.WithInstrumentationAttributes(col.labels...))
		opt.Filters = []string{
			"timestamp gt \"" + ctime.Format("2006-01-02T15:04:05.000Z") + "\"",
		}

		tmpTime := time.Now().UTC()
		data, err := client.GetInstances(&opt)
		if err != nil {
			logger.Error("Error to GET AlertLog", "err", err)
			time.Sleep(col.interval)
			continue
		}
		if len(data) == 0 {
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
				Message   string `json:"message"`
				MessageId string `json:"message_id"`
			}{
				v.Get("message").String(),
				v.Get("messageId").String(),
			}
			record.SetTimestamp(v.Get("timestamp").Time())
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
