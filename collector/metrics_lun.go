package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
)

func collectMetricLun(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()
	var cols collectorSt
	cols.subName = "lun"
	cols.apiPath = "/api/types/metricValue/instances"

	metrics := []string{
		"sp.*.storage.lun.*.readBytesRate",
		"sp.*.storage.lun.*.readsRate",
		"sp.*.storage.lun.*.writeBytesRate",
		"sp.*.storage.lun.*.writesRate",
		"sp.*.storage.lun.*.queueLength",
		"sp.*.storage.lun.*.responseTime",
	}

	cols.convMetric2GaugeVec(metrics)
	cols.getMetric(uc, reg)

	return true
}
