package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
	"unisphere_exporter/utils"
)

func collectLun(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	var result float64

	cols := utils.NonMetricCollector{
		ApiPath:   "/api/types/lun/instances",
		Namespace: namespace,
		SubName:   "lun",
		Registry:  reg,
		Labels:    []string{"lun", "lun_name", "pool"},
	}

	cols.CreateGaugeVec(types.LunContent{})

	var jData types.LunEntries
	if !cols.GetGaugeValue(uc, &jData) {
		return result
	}

	if jData.Entries == nil {
		uc.Logger.Error("Contents is Null.", "subsystem", cols.SubName)
		return result
	}
	for _, entry := range jData.Entries {
		content := entry.Content
		for k, _ := range cols.MetricGaugeVecList {
			v := reflect.ValueOf(content).FieldByName(k)
			cols.MetricGaugeVecList[k].WithLabelValues(content.Id, content.Name, content.Pool.Id).Set(utils.Types2Float64(v))
		}
	}

	result = 1.0
	return result
}
