package main

import (
	"log/slog"
	"os"
	"unisphere_exporter/config"
	"unisphere_exporter/provider"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	promslogflag "github.com/prometheus/common/promslog/flag"
)

var (
	configFile = kingpin.Flag("config.file", "Path to config file.").Short('c').Default("config.file").String()
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

	if !provider.RegistryProviders(cfg, logger) {
		isFailed = true
	}

	if isFailed {
		logger.Error("Failed to load configs...")
		os.Exit(1)
	}

	provider.RunProviders(logger)
}
