package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

// import (
//
//	"github.com/prometheus/client_golang/prometheus"
//	"strings"
//	"sync"
//	"unisphere_exporter/utils"
//
// )

func init() {
	NewCollector(NewMetricFCCollector())
}

func NewMetricFCCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime_fc"
	m.metricPath = []string{
		"sp.*.fibreChannel.fePort.*.readBlocks",
		"sp.*.fibreChannel.fePort.*.readBytesRate",
		"sp.*.fibreChannel.fePort.*.reads",
		"sp.*.fibreChannel.fePort.*.readsRate",
		"sp.*.fibreChannel.fePort.*.writeBlocks",
		"sp.*.fibreChannel.fePort.*.writeBytesRate",
		"sp.*.fibreChannel.fePort.*.writes",
		"sp.*.fibreChannel.fePort.*.writesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}

func (c *MetricCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var result float64
	var mq types.MetricQueryEntries

	qid := uc.PostMetricRealTimeQuery(c.metricPath, 60)
	if qid == 0 {
		return result
	}
	uc.GetMetricRealTimeQueryResult(qid)
	for _, entry := range mq.Entries {
		content := entry.Content
		//ch <- prometheus.MustNewConstMetric(metricList[content.Path], prometheus.GaugeValue, 1.0, "123")
		log.Println(c.metricList[content.Path])

	}
	result = 1.0
	return result
}

//
//func ProbeMetricFC(c *utils.UnisphereClient, registry *prometheus.Registry, wg *sync.WaitGroup) float64 {
//	// Variable qr is return value
//	defer wg.Done()
//	qr := 0.0
//	labels := []string{"sp_name", "fc_name"}
//	interval := 60
//
//	paths := []string{
//		"sp.*.fibreChannel.fePort.*.readBlocks",
//		"sp.*.fibreChannel.fePort.*.readBytesRate",
//		"sp.*.fibreChannel.fePort.*.reads",
//		"sp.*.fibreChannel.fePort.*.readsRate",
//		"sp.*.fibreChannel.fePort.*.writeBlocks",
//		"sp.*.fibreChannel.fePort.*.writeBytesRate",
//		"sp.*.fibreChannel.fePort.*.writes",
//		"sp.*.fibreChannel.fePort.*.writesRate",
//	}
//	qid, qr := MetricQuery(c, paths, interval)
//	if qr == 0.0 {
//		return qr
//	}
//	qresult, qr := MetricQueryResult(c, qid)
//	if qr == 0.0 {
//		return qr
//	}
//
//	for _, entry := range qresult.Entries {
//		p := strings.ReplaceAll(entry.Content.Path, "sp.*.", "")
//		p = strings.ReplaceAll(p, ".*.", "_")
//		p = strings.ReplaceAll(p, ".", "_")
//		p = strings.ToLower("unisphere_" + p)
//
//		mSpMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: p, Help: "Full Path is " + entry.Content.Path}, labels)
//		registry.Register(mSpMetric)
//
//		for k1, v1 := range entry.Content.Values {
//			var f float64
//
//			switch i := v1.(type) {
//			case int:
//				f = float64(i)
//			case float32:
//				f = float64(i)
//			case float64:
//				f = i
//			case interface{}:
//				for k2, v2 := range v1.(map[string]interface{}) {
//					switch j := v2.(type) {
//					case int:
//						f = float64(j)
//					case float32:
//						f = float64(j)
//					case float64:
//						f = j
//					case bool:
//						if j == true {
//							f = 1.0
//						} else {
//							f = 0.0
//						}
//					default:
//						return 0.0
//					}
//					mSpMetric.WithLabelValues(k1, k2).Set(f)
//				}
//			case bool:
//				if i == true {
//					f = 1.0
//				} else {
//					f = 0.0
//				}
//			default:
//				return 0.0
//			}
//		}
//	}
//
//	return 1.0
//}
