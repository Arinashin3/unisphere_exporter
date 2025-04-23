package main

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"net/http"
	"os/user"
	"runtime"
	"unisphere_exporter/client"
	"unisphere_exporter/collector"
)

func main() {
	var (
		configFile = kingpin.Flag("config.file", "file containing the authentication map to use when connecting to a Unisphere device").Default("config.yml").String()
		listen     = kingpin.Flag("web.listen-address", "Addresses on which to expose metrics and web interface.").Default(":9182").String()
		maxProcs   = kingpin.Flag("runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)").Envar("GOMAXPROCS").Default("1").Int()
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("unisphere_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	logger.Info("Starting unisphere_exporter", "version", version.Info())
	logger.Info("Build contex", "build_context", version.BuildContext())

	if u, err := user.Current(); err == nil && u.Uid == "0" {
		logger.Warn("Unisphere Exporter is running as root user. This exporter is designed to run as unprivileged user, root is not required.")
	}
	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	client.SetModule(configFile, logger)

	logger.Info("Unisphere exporter running.", "listen_port", *listen)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewRegistry()
		collector.Probe(w, r, logger, reg)
	})

	http.ListenAndServe(*listen, nil)

}
