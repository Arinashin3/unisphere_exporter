package utils

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

type MetricCollector struct {
	Namespace          string
	SubName            string
	ApiPath            string
	MetricPath         []string
	Registry           *prometheus.Registry
	MetricGaugeVecList map[string]*prometheus.GaugeVec
}

func (c *MetricCollector) CreateGaugeVec() {

	c.MetricGaugeVecList = make(map[string]*prometheus.GaugeVec)

	for _, path := range c.MetricPath {
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
		c.MetricGaugeVecList[path] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: c.Namespace, Subsystem: c.SubName, Name: name}, labels)
		c.Registry.MustRegister(c.MetricGaugeVecList[path])
	}
}
func (c *MetricCollector) GetGaugeValue(uc *client.UnisphereClient) bool {
	ok := true
	logger := uc.Logger

	for k, _ := range c.MetricGaugeVecList {
		var jData types.MetricQueryEntries
		query := "filter=path%20EQ%20\"" + k + "\""
		query += "&compact=true"
		data := uc.Send("GET", c.ApiPath, query, nil)
		if data == nil {
			logger.Error("Data is Nil.", "path", k)
			ok = false
			continue
		}
		if json.Unmarshal(data, &jData) != nil {
			logger.Error("Unmarshal is Failed", "path", k)
			ok = false
			continue
		}
		if jData.Entries == nil {
			logger.Error("Entries is Null", "path", k)
			ok = false
			continue
		}

		logger.Debug("Success Request and Parsing to Json", "subsystem", c.SubName, "metric", k)
		c.generateMetrics(k, jData.Entries[0].Content.Values)

	}
	return ok
}

// generateMetrics
// MetricValue의 추출 값을 파싱한다.
// 2회전까지만 사용 가능.
func (c *MetricCollector) generateMetrics(path string, value map[string]interface{}) bool {
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
					c.MetricGaugeVecList[path].WithLabelValues(k1, k2).Set(Types2Float64(val2))
				}
			}
		} else if strings.Contains(val1.Type().String(), "string") {
			return result
		} else {
			c.MetricGaugeVecList[path].WithLabelValues(k1).Set(Types2Float64(val1))
		}
	}

	result = true
	return result

}
