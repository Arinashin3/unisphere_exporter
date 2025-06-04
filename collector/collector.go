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

	uc := client.LoadClient(target, module, logger)
	if uc == nil {
		return
	}

	// Connection Test
	resp := uc.Send("GET", "/api/types/loginSessionInfo/instances", "compact=true", nil)
	if resp == nil {
		logger.Error("Connection Failed.", "client", target, "module", module)
		return
	}
	reg := prometheus.NewRegistry()
	var wg sync.WaitGroup
	wg.Add(10)
	go collectBasicSystemInfo(uc, reg, &wg)
	go collectPool(uc, reg, &wg)
	go collectSystemCapacity(uc, reg, &wg)
	go collectMetricFC(uc, reg, &wg)
	go collectMetricGlobal(uc, reg, &wg)
	go collectMetricLun(uc, reg, &wg)
	go collectLun(uc, reg, &wg)
	go collectMetricFilesystem(uc, reg, &wg)
	go collectFilesystem(uc, reg, &wg)
	go collectReplication(uc, reg, &wg)
	wg.Wait()
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(*w, r)

}
