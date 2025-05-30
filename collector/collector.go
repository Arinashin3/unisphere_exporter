package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"sync"
	"unisphere_exporter/client"
)

const namespace = "unisphere"

var ucList map[string]*client.UnisphereClient

func init() {
	ucList = make(map[string]*client.UnisphereClient)

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

	var connected bool
	if ucList[target] == nil {
		ucList[target], connected = client.CreateClient(target, module, logger)

		if !connected {
			return
		}
	}

	uc := ucList[target]
	uc.ClientLogIn()

	//uc, connected := client.NewClient(target, module, logger)
	reg := prometheus.NewRegistry()
	var wg sync.WaitGroup
	wg.Add(9)
	go collectBasicSystemInfo(uc, reg, &wg)
	go collectPool(uc, reg, &wg)
	go collectSystemCapacity(uc, reg, &wg)
	go collectMetricFC(uc, reg, &wg)
	go collectMetricGlobal(uc, reg, &wg)
	go collectMetricLun(uc, reg, &wg)
	go collectLun(uc, reg, &wg)
	go collectMetricFilesystem(uc, reg, &wg)
	go collectFilesystem(uc, reg, &wg)
	wg.Wait()
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(*w, r)

}
