package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"unisphere_exporter/client"
	"unisphere_exporter/utils"
)

func init() {
	NewCollector("metric", NewMetric())
	//enabled := true
	//var test BasicSystemInfoCollector
	//
	//var m MetricSt
	//descField := reflect.TypeOf(desc)
	//for i := 0; i < descField.NumField(); i++ {
	//	//fn := descField.Field(i).Name
	//
	//	//fqname := NewFQName(subsystem, fn)
	//
	//}
	//log.Println(c)
	//for i, i := range  {

	//}
}
func NewMetric() Collector {
	return &MetricCollector{}
}

type MetricCollector struct {
	infoDesc *prometheus.Desc
}

func (c *MetricCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {

	return 0.0
}
func ProbeMetric(c utils.UnisphereHTTP, registry *prometheus.Registry) bool {
	var mStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "unisphere_cpu_test", Help: "Cumulative CPU usage in percent"}, []string{"status"})
	registry.MustRegister(mStatus)

	type Content struct {
		Id                  int    `json:"id"`
		Name                string `json:"name"`
		Path                string `json:"path"`
		IsRealtimeAvailable bool   `json:"isRealtimeAvailable"`
	}
	type Entries []struct {
		Content `json:"content"`
	}
	type AllMetric struct {
		Entries `json:"entries"`
	}

	var st AllMetric
	if err := c.Get("/api/types/metric/instances", "compact=true", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st.Entries {
		go func() {
			println(s.Id)
			println(s.Name)
			println(s.Path)
		}()
	}
	mStatus.WithLabelValues("online").Set(1)
	return true
}
