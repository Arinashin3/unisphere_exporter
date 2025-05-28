package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func (cols *CollectSt) getMetric(uc *client.UnisphereClient, reg *prometheus.Registry) {

	for k, _ := range cols.metricList {
		var jData types.MetricQueryEntries
		query := "filter=path%20EQ%20\"" + k + "\""
		query += "&compact=true"
		data := uc.Get(cols.apiPath, query)
		if json.Unmarshal(data, &jData) != nil {
			uc.Logger.Error("Unmarshal is Failed", "path", k)
			continue
		}
		if jData.Entries == nil {
			uc.Logger.Error("Entries is Null", "path", k)
			continue
		}
		reg.MustRegister(cols.metricList[k])

		cols.generateMetrics(k, jData.Entries[0].Content.Values)

	}
}

func (cols *CollectSt) convMetric2GaugeVec(metrics []string) bool {

	cols.metricList = make(map[string]*prometheus.GaugeVec)

	for _, path := range metrics {
		var labels []string
		var name string
		arr := strings.Split(path, ".")
		for i := 0; i < len(arr); i++ {
			if i == 0 {
				// 배열 첫번째인지 확인
				continue
			}
			// 문자 * 인지 확인, 맞을 경우 라벨에 포함, 아닐경우 이전 값이 * 인지 확인해서 이름에 포함
			if arr[i] == "*" {
				labels = append(labels, arr[i-1])
			} else if arr[i-1] != "*" {
				if name == "" {
					name = arr[i-1]

				} else {
					name = name + "_" + arr[i-1]
				}
			}
			if i == len(arr)-1 {
				// 배열 마지막인지 확인
				name = name + "_" + arr[i]
				continue
			}

		}

		cols.metricList[path] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: Namespace, Subsystem: cols.subName, Name: name}, labels)

	}
	return true

}

func convValue2Float64(value reflect.Value) float64 {
	v := value.Type().String()

	switch {
	case strings.Contains(v, "uint"):
		return float64(value.Uint())
	case strings.Contains(v, "float"):
		return value.Float()
	case strings.Contains(v, "int"):
		return float64(value.Int())
	default:
		return 0
	}

}

func (cols *CollectSt) generateMetrics(path string, value map[string]interface{}) bool {
	var result bool

	for k1, v1 := range value {
		val1 := reflect.ValueOf(v1)
		if strings.Contains(val1.Type().String(), "interface") {
			for k2, v2 := range v1.(map[string]interface{}) {
				val2 := reflect.ValueOf(v2)
				if strings.Contains(val2.Type().String(), "interface") {
					return result
				} else if strings.Contains(val2.Type().String(), "string") {
					return result
				} else {
					cols.metricList[path].WithLabelValues(k1, k2).Set(convValue2Float64(val2))
				}
			}
		} else if strings.Contains(val1.Type().String(), "string") {
			return result
		} else {
			cols.metricList[path].WithLabelValues(k1).Set(convValue2Float64(val1))
		}
	}

	result = true
	return result

}
