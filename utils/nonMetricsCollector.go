package utils

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"reflect"
	"strings"
	"unisphere_exporter/client"
)

// NonMetricCollector
// API Path : 'MetricValue'가 아닐 것!!
// 필수 요소 : Namespace + SubName + ApiPath + Labels + Registry
type NonMetricCollector struct {
	Namespace          string
	SubName            string
	ApiPath            string
	Labels             []string
	Registry           *prometheus.Registry
	Logger             *slog.Logger
	requiredFields     []string
	MetricGaugeVecList map[string]*prometheus.GaugeVec
}

// CreateGaugeVec
// GaugeVec 자동 생성기
// 구조체에 Int, Uint, Float 타입이 필수로 있을 것!
func (c *NonMetricCollector) CreateGaugeVec(i interface{}) {
	c.MetricGaugeVecList = make(map[string]*prometheus.GaugeVec)
	for _, field := range reflect.VisibleFields(reflect.TypeOf(i)) {
		fType := field.Type.String()
		fName := strings.Trim(string(field.Tag), "json:")
		fName = strings.Trim(fName, "\"")
		c.requiredFields = append(c.requiredFields, fName)
		if fType == "string" {
			continue
		} else if strings.Contains(fType, "struct") {
			continue
		} else {
			c.MetricGaugeVecList[field.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: c.Namespace, Subsystem: c.SubName, Name: field.Name}, c.Labels)
			c.Registry.MustRegister(c.MetricGaugeVecList[field.Name])
		}
	}
}

// GetGaugeValue
// 클라이언트로부터 데이터 요청(GET 방식) 및 Json 파싱
func (c *NonMetricCollector) GetGaugeValue(uc *client.UnisphereClient, outputJson any) bool {
	query := "fields=" + strings.Join(c.requiredFields, ",")
	query += "&compact=true"

	if json.Unmarshal(uc.Get(c.ApiPath, query), outputJson) != nil {
		c.Logger.Error("Unmarshal is Failed.", "subsystem", c.SubName)
		return false
	}

	return true
}
