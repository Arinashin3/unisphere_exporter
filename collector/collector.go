package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"reflect"
	"sync"
	"unisphere_exporter/client"
)

const namespace = "unisphere"

type collectorSt struct {
	subName    string
	apiPath    string
	labels     []string
	metricList map[string]*prometheus.GaugeVec
}

func Probe(w *http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	params := r.URL.Query()
	target := params.Get("target")
	module := params.Get("module")

	if target == "" {
		logger.Error("Target parameter is empty.")
		return
	} else if module == "" {
		logger.Error("Module parameter is empty.")
		return
	}

	uc, connected := client.NewClient(target, module, logger)
	if !connected {
		return
	}
	reg := prometheus.NewRegistry()
	var wg sync.WaitGroup
	wg.Add(6)
	go collectBasicSystemInfo(uc, reg, &wg)
	go collectPool(uc, reg, &wg)
	go collectSystemCapacity(uc, reg, &wg)
	go collectMetricFC(uc, reg, &wg)
	go collectMetricGlobal(uc, reg, &wg)
	go collectMetricLun(uc, reg, &wg)
	wg.Wait()
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(*w, r)

}

func types2Float64(i reflect.Value) float64 {
	switch i.Type().String() {
	case "int":
		return float64(i.Int())
	case "uint64":
		return float64(i.Uint())
	case "float64":
		return i.Float()
	case "bool":
		if i.Bool() {
			return 1.0
		}
	}
	return 0.0
}
