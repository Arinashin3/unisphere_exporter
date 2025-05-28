package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func collectBasicSystemInfo(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()

	var result float64
	var cols CollectSt
	cols.subName = "basicsystem"
	cols.apiPath = "/api/types/basicSystemInfo/instances"
	cols.labels = []string{"model", "sw_ver", "api_ver"}
	cols.metricList = make(map[string]*prometheus.GaugeVec)
	cols.metricList["info"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: Namespace, Subsystem: cols.subName, Name: "info", Help: ""}, cols.labels)

	for k, _ := range cols.metricList {
		reg.MustRegister(cols.metricList[k])
	}

	var jData types.BasicSystemInfoEntries
	resp := uc.Get(cols.apiPath, "compact=true")
	if resp == nil {
		uc.Logger.Error("Data is Null.", "subsystem", cols.subName)
		return result
	}
	if json.Unmarshal(resp, &jData) != nil {
		uc.Logger.Error("Unmarshal Failed.", "subsystem", cols.subName)
		return result
	}
	if jData.Entries == nil {
		uc.Logger.Error("Contents is Null.", "subsystem", cols.subName)
		return result
	}
	content := jData.Entries[0].Content
	cols.metricList["info"].WithLabelValues(content.Model, content.SoftwareFullVersion, content.ApiVersion).Set(0)

	result = 1.0
	return result
}
