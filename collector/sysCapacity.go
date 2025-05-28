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

func collectSystemCapacity(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	var result float64
	var cols CollectSt
	var requiredFields []string
	cols.subName = "syscap"
	cols.apiPath = "/api/types/systemCapacity/instances"
	cols.metricList = make(map[string]*prometheus.GaugeVec)
	cols.labels = []string{}

	// 메트릭 리스트 생성
	for _, f := range reflect.VisibleFields(reflect.TypeOf(types.SysCapContent{})) {
		fType := f.Type.String()
		fName := strings.Trim(string(f.Tag), "json:")
		fName = strings.Trim(fName, "\"")
		requiredFields = append(requiredFields, fName)
		if fType != "string" {
			cols.metricList[f.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: Namespace, Subsystem: cols.subName, Name: f.Name}, cols.labels)
		}
	}
	query := "fields=" + strings.Join(requiredFields, ",")
	query += "&compact=true"

	// Registry 등록
	for k, _ := range cols.metricList {
		reg.MustRegister(cols.metricList[k])
	}

	// Data 요청
	var jData types.SysCapEntries
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
		cols.metricList[k].WithLabelValues().Set(Types2Float64(v))
	}

	result = 1.0
	return result
}
