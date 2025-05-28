package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
)

func collectMetricFC(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) bool {
	defer wg.Done()
	var cols CollectSt
	cols.subName = "fc"
	cols.apiPath = "/api/types/metricValue/instances"

	metrics := []string{
		"sp.*.fibreChannel.fePort.*.readBytesRate",
		"sp.*.fibreChannel.fePort.*.readsRate",
		"sp.*.fibreChannel.fePort.*.writeBytesRate",
		"sp.*.fibreChannel.fePort.*.writesRate",
	}

	cols.convMetric2GaugeVec(metrics)
	cols.getMetric(uc, reg)

	return true
}
