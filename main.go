package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
	"unisphere_exporter/collectors"
	"unisphere_exporter/config"
	"unisphere_exporter/gounity"
	"unisphere_exporter/utils"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	promslogflag "github.com/prometheus/common/promslog/flag"
	"go.opentelemetry.io/otel/attribute"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

const serviceName = "unisphere_exporter"

var (
	configFile = kingpin.Flag("config.file", "Paths to config file.").Short('c').Default("config.yml").String()
	logger     *slog.Logger
	isFailed   bool
)

func main() {
	// Set Flag & Logger
	promslogConfig := &promslog.Config{}
	promslogflag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger = promslog.New(promslogConfig)

	// Load Configuration Set Configurations...
	logger.Info("Load Configs...")
	var cfg *config.UnisphereConfig
	cfg = config.NewConfiguration()
	err := cfg.LoadFile(configFile)
	if err != nil {
		isFailed = true
		logger.Error("Failed to load config file.", "error", err)
	}

	// Set Config Providers...
	//for k, v := range cfg.Providers {
	// Check Providers...
	//if !provider.FindDefaultProvider(k) {
	//	continue
	//}
	//if v == nil {
	//provider.EnabledProvider[k] = true
	//continue
	//}

	// Check Provider Enabled
	// When Provider is existed in config file,
	// Allow Provider setting in config file.
	//enabledConf := v.(map[string]interface{})["enabled"]
	//if enabledConf != nil {
	//enabled := enabledConf.(bool)
	//if provider.FindDefaultProvider(k) {
	//provider.EnabledProvider[k] = enabled
	//}
	//} else {
	//provider.EnabledProvider[k] = true
	//}

	//}

	// Create Exporters...
	var me *sdkMetric.Exporter
	var le *sdkLog.Exporter
	ctx := context.Background()
	mConf := cfg.Server.Metrics
	insecure, _ := strconv.ParseBool(mConf.Insecure)
	if me, err = utils.NewMetricExporter(ctx, mConf.Mode, mConf.Endpoint+mConf.Api_Path, insecure); err != nil {
		logger.Error("Failed to create metric exporter.", "error", err)
		isFailed = true
	}

	lConf := cfg.Server.Logs
	insecure, _ = strconv.ParseBool(lConf.Insecure)
	if le, err = utils.NewLogExporter(ctx, lConf.Mode, lConf.Endpoint+lConf.Api_Path, insecure); err != nil {
		logger.Error("Failed to create log exporter.", "error", err)
		isFailed = true
	}

	// Create Collectors...
	var cols = make(map[string]*collectors.Collector)
	var insecureTr = gounity.NewTransport(true)
	var secureTr = gounity.NewTransport(false)
	for _, clientConf := range cfg.Clients {
		var attrs []attribute.KeyValue
		for k, v := range clientConf.Labels {
			attrs = append(attrs, attribute.String(k, v))
		}
		attrs = append(attrs, attribute.String("instance", clientConf.Endpoint))
		attrs = append(attrs, attribute.String("service.name", serviceName))

		// Create Collector...
		var col *collectors.Collector
		if col, err = collectors.NewCollector(ctx, attrs); err != nil {
			logger.Error("Failed to create collector.", "error", err)
			isFailed = true
		}

		// Set Collector's Client...
		username, password := cfg.SearchAuth(clientConf.Auth)
		if username == "" || password == "" {
			logger.Error("Failed to get username and password.")
			isFailed = true
		}
		var tr *http.Transport
		if insecure, _ = strconv.ParseBool(clientConf.Insecure); insecure {
			tr = insecureTr
		} else {
			tr = secureTr
		}
		col.Client = gounity.NewUnisphereClient(clientConf.Endpoint, username, password, tr)

		// Set Collector's Provider...
		var interval time.Duration
		if interval, err = time.ParseDuration(clientConf.Interval); err != nil {
			logger.Error("Failed to parse interval.", "error", err)
			isFailed = true
		}
		col.SetInterval(interval)
		col.NewMeterProvider(interval, me)
		col.NewLoggerProvider(interval, le)

		cols[clientConf.Endpoint] = col
	}

	if isFailed {
		logger.Error("Failed to load configs...")
		os.Exit(1)
	}

	for _, col := range cols {
		go col.Start(logger)
	}
	select {}
}
