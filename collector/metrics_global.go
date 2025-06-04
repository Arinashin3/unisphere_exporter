package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/utils"
)

func collectMetricGlobal(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()

	cols := utils.MetricCollector{
		Namespace: namespace,
		SubName:   "global",
		ApiPath:   "/api/types/metricValue/instances",
		Registry:  reg,
	}
	cols.MetricPath = []string{
		"sp.*.cpu.summary.utilization",
		"sp.*.blockCache.global.summary.dirtyBytes",
		"sp.*.net.basic.inBytesRate",
		"sp.*.net.basic.outBytesRate",
	}

	cols.CreateGaugeVec()
	cols.GetGaugeValue(uc)

	return true
}
