package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
)

func collectMetricGlobal(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()
	var cols CollectSt
	cols.subName = "global"
	cols.apiPath = "/api/types/metricValue/instances"

	metrics := []string{
		"sp.*.cpu.summary.utilization",
		"sp.*.blockCache.global.summary.dirtyBytes",
		"sp.*.net.basic.inBytesRate",
		"sp.*.net.basic.outBytesRate",
	}

	cols.convMetric2GaugeVec(metrics)
	cols.getMetric(uc, reg)

	return true
}
