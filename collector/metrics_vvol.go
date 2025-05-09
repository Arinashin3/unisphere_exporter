package collector

//import (
//	"github.com/prometheus/client_golang/prometheus"
//	"strings"
//
//	//"strings"
//	"sync"
//	"unisphere_exporter/utils"
//)
//
//func ProbeMetricVVol(c utils.UnisphereClient, registry *prometheus.Registry, wg *sync.WaitGroup) float64 {
//	// Variable qr is return value
//	defer wg.Done()
//	qr := 0.0
//	labels := []string{"sp_name", "volume_name"}
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
//
//	qid, qr := MetricQuery(&c, paths, interval)
//	if qr == 0.0 {
//		return qr
//	}
//	qresult, qr := MetricQueryResult(&c, qid)
//	if qr == 0.0 {
//		return qr
//	}
//
//	for _, entry := range qresult.Entries {
//		p := strings.ReplaceAll(entry.Content.Path, "sp.*.", "")
//		p = strings.ReplaceAll(p, ".*.", "_")
//		p = strings.ReplaceAll(p, ".", "_")
//		p = strings.ToLower("unisphere_metric_" + p)
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
