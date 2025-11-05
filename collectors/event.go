package collectors

import (
	"encoding/json"
	"log/slog"
	"time"
	"unisphere_exporter/gounity/api"
	"unisphere_exporter/utils/enum"

	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/log"
	"gopkg.in/yaml.v3"
)

func init() {
	key := "event"
	RegisterModule(key, NewEvent())
}

type ModuleEvent struct {
	// Module's Information
	name      string
	opts      *api.UnityActionOptions
	defaults  bool // Default Enabled
	timestamp time.Time

	// Configuration File
	Enabled *bool `yaml:"enabled"`
	Level   int64 `yaml:"level"`
}

func NewEvent() *ModuleEvent {
	return &ModuleEvent{
		defaults: false,
		Level:    5,
	}
}

func (_m *ModuleEvent) GetEnabled() bool {
	return *_m.Enabled
}

func (_m *ModuleEvent) SetConfig(body []byte) {
	err := yaml.Unmarshal(body, _m)
	if err != nil {
		panic(err)
	}
	if _m.Enabled == nil {
		_m.Enabled = &_m.defaults
	}
}

func (_m *ModuleEvent) Init(key string) {
	_m.name = key
	_m.opts = api.NewUnityActionOptions("event")
	_m.opts.Fields = []string{"creationTime", "severity", "messageId", "message", "source"}
}

func (_m *ModuleEvent) Run(logger *slog.Logger, col *Collector) {
	opt := *_m.opts
	ctime := time.Now().Add(-1 * time.Hour).UTC()
	client := col.Client
	lp := col.loggerProvider

	for {
		pvlogger := lp.Logger(_m.name, log.WithInstrumentationAttributes(col.labels...))
		opt.Filters = []string{
			"creationTime gt \"" + ctime.Format("2006-01-02T15:04:05.000Z") + "\"",
		}

		tmpTime := time.Now().UTC()
		data, err := client.GetInstances(&opt)
		if err != nil {
			logger.Error("Error to GET EventLog", "err", err)
			col.success = false
			time.Sleep(col.interval)
			continue
		}
		col.success = true
		if data == nil {
			time.Sleep(col.interval)
			continue
		}

		for _, v := range data {
			record := log.Record{}
			if _m.Level > v.Get("severity").Int() {
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
