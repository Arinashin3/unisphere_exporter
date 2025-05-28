package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"sync"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func collectLun(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	var result float64
	var cols collectorSt
	var requiredFields []string
	cols.subName = "lun"
	cols.apiPath = "/api/types/lun/instances"
	cols.metricList = make(map[string]*prometheus.GaugeVec)
	cols.labels = []string{"lun_id", "lun_name"}

	// 메트릭 리스트 생성
	for _, f := range reflect.VisibleFields(reflect.TypeOf(types.LunContent{})) {
		fType := f.Type.String()
		fName := strings.Trim(string(f.Tag), "json:")
		fName = strings.Trim(fName, "\"")
		if fType != "string" {
			cols.metricList[f.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: namespace, Subsystem: cols.subName, Name: f.Name}, cols.labels)
		}
	}
	query := "fields=" + strings.Join(requiredFields, ",")
	query += "&compact=true"
	// Registry 등록
	for k, _ := range cols.metricList {
		reg.MustRegister(cols.metricList[k])
	}

	// Data 요청
	var jData types.LunEntries
	resp := uc.Get(cols.apiPath, query)
	if json.Unmarshal(resp, &jData) != nil {
		uc.Logger.Error("Unmarshal Failed.", "subsystem", cols.subName)
		return result
	}
	if jData.Entries == nil {
		uc.Logger.Error("Contents is Null.", "subsystem", cols.subName)
		return result
	}
	content := jData.Entries[0].Content
	for k, _ := range cols.metricList {
		v := reflect.ValueOf(content).FieldByName(k)
		cols.metricList[k].WithLabelValues(content.Id, content.Name).Set(types2Float64(v))
	}

	result = 1.0
	return result
}
