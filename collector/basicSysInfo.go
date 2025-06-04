package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
	"unisphere_exporter/utils"
)

func collectBasicSystemInfo(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()

	var result float64

	cols := utils.NonMetricCollector{
		ApiPath:   "/api/types/basicSystemInfo/instances",
		Namespace: namespace,
		SubName:   "basicsystem",
		Registry:  reg,
		Labels:    []string{"model", "sw_ver", "api_ver"},
	}

	cols.MetricGaugeVecList = make(map[string]*prometheus.GaugeVec)
	cols.MetricGaugeVecList["info"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: namespace, Subsystem: cols.SubName, Name: "info", Help: ""}, cols.Labels)
	cols.Registry.MustRegister(cols.MetricGaugeVecList["info"])

	var jData types.BasicSystemInfoEntries
	if !cols.GetGaugeValue(uc, &jData) {
		return result
	}
	if jData.Entries == nil {
		uc.Logger.Error("Contents is Null.", "subsystem", cols.SubName)
		return result
	}
	for _, entry := range jData.Entries {
		content := entry.Content
		for range cols.MetricGaugeVecList {
			cols.MetricGaugeVecList["info"].WithLabelValues(content.Model, content.SoftwareFullVersion, content.ApiVersion).Set(1)
		}
	}

	result = 1.0
	return result
}
