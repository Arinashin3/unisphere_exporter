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

func collectPool(uc *client.UnisphereClient, reg *prometheus.Registry, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	var result float64
	var cols CollectSt
	cols.subName = "pool"
	cols.apiPath = "/api/types/pool/instances"
	cols.metricList = make(map[string]*prometheus.GaugeVec)
	cols.labels = []string{"pool_id", "pool_name"}

	query := "fields=id,name,raidType,sizeFree,sizeTotal,sizeUsed"
	query += "&compact=true"
	// 메트릭 리스트 생성
	for _, f := range reflect.VisibleFields(reflect.TypeOf(types.PoolContent{})) {
		fType := f.Type.String()
		fName := strings.Trim(string(f.Tag), "json:")
		fName = strings.Trim(fName, "\"")
		if fType != "string" {
			cols.metricList[f.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: Namespace, Subsystem: cols.subName, Name: f.Name}, cols.labels)
		}
	}

	// Registry 등록
	for k, _ := range cols.metricList {
		reg.MustRegister(cols.metricList[k])
	}

	// Data 요청
	var jData types.PoolEntries
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
		cols.metricList[k].WithLabelValues(content.ID, content.Name).Set(Types2Float64(v))
	}

	result = 1.0
	return result
}
