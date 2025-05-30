package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/utils"
)

func collectMetricFilesystem(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()
	cols := utils.MetricCollector{
		Namespace: namespace,
		SubName:   "fs",
		ApiPath:   "/api/types/metricValue/instances",
		Registry:  reg,
		Logger:    uc.Logger,
	}
	cols.MetricPath = []string{
		"sp.*.storage.filesystem.*.readBytesRate",
		"sp.*.storage.filesystem.*.readsRate",
		"sp.*.storage.filesystem.*.writeBytesRate",
		"sp.*.storage.filesystem.*.writesRate",
	}
	cols.CreateGaugeVec()
	cols.GetGaugeValue(uc)

	return true
}
