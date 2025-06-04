package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/utils"
)

func collectMetricLun(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()

	cols := utils.MetricCollector{
		Namespace: namespace,
		SubName:   "lun",
		ApiPath:   "/api/types/metricValue/instances",
		Registry:  reg,
	}
	cols.MetricPath = []string{
		"sp.*.storage.lun.*.readBytesRate",
		"sp.*.storage.lun.*.readsRate",
		"sp.*.storage.lun.*.writeBytesRate",
		"sp.*.storage.lun.*.writesRate",
		"sp.*.storage.lun.*.queueLength",
		"sp.*.storage.lun.*.responseTime",
	}

	cols.CreateGaugeVec()
	cols.GetGaugeValue(uc)

	return true
}
