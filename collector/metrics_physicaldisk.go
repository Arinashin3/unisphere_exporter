package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
	"unisphere_exporter/utils"
)

func ProbeMetricPhysicalDisk(c utils.UnisphereClient, registry *prometheus.Registry, wg *sync.WaitGroup) float64 {
	// Variable qr is return value
	defer wg.Done()
	qr := 0.0
	labels := []string{"sp_name", "disk_name"}
	interval := 60

	paths := []string{
		"sp.*.physical.disk.*.averageQueueLength",
		"sp.*.physical.disk.*.busyTicks",
		"sp.*.physical.disk.*.idleTicks",
		"sp.*.physical.disk.*.readBlocks",
		"sp.*.physical.disk.*.readBytesRate",
		"sp.*.physical.disk.*.reads",
		"sp.*.physical.disk.*.readsRate",
		"sp.*.physical.disk.*.responseTime",
		"sp.*.physical.disk.*.serviceTime",
		"sp.*.physical.disk.*.sumArrivalQueueLength",
		"sp.*.physical.disk.*.totalCallsRate",
		"sp.*.physical.disk.*.writeBlocks",
		"sp.*.physical.disk.*.writeBytesRate",
		"sp.*.physical.disk.*.writes",
		"sp.*.physical.disk.*.writesRate",
	}

	qid, qr := MetricQuery(&c, paths, interval)
	if qr == 0.0 {
		return qr
	}
	qresult, qr := MetricQueryResult(&c, qid)
	if qr == 0.0 {
		return qr
	}

	for _, entry := range qresult.Entries {
		p := strings.ReplaceAll(entry.Content.Path, "sp.*.", "")
		p = strings.ReplaceAll(p, ".*.", "_")
		p = strings.ReplaceAll(p, ".", "_")
		p = strings.ToLower("unisphere_" + p)

		mSpMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: p, Help: "Full Path is " + entry.Content.Path}, labels)
		registry.Register(mSpMetric)

		for k1, v1 := range entry.Content.Values {
			var f float64

			switch i := v1.(type) {
			case int:
				f = float64(i)
			case float32:
				f = float64(i)
			case float64:
				f = i
			case interface{}:
				for k2, v2 := range v1.(map[string]interface{}) {
					switch j := v2.(type) {
					case int:
						f = float64(j)
					case float32:
						f = float64(j)
					case float64:
						f = j
					case bool:
						if j == true {
							f = 1.0
						} else {
							f = 0.0
						}
					default:
						return 0.0
					}
					mSpMetric.WithLabelValues(k1, k2).Set(f)
				}
			case bool:
				if i == true {
					f = 1.0
				} else {
					f = 0.0
				}
			default:
				return 0.0
			}
		}
	}

	return 1.0
}
