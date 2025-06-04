package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/utils"
)

func collectMetricFC(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()
	cols := utils.MetricCollector{
		Namespace: namespace,
		SubName:   "fc",
		ApiPath:   "/api/types/metricValue/instances",
		Registry:  reg,
	}
	cols.MetricPath = []string{
		"sp.*.fibreChannel.fePort.*.readBytesRate",
		"sp.*.fibreChannel.fePort.*.readsRate",
		"sp.*.fibreChannel.fePort.*.writeBytesRate",
		"sp.*.fibreChannel.fePort.*.writesRate",
	}
	cols.CreateGaugeVec()
	cols.GetGaugeValue(uc)

	return true
}
