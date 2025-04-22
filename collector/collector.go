package collector

import (
	//"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"time"
	"unisphere_exporter/client"

	//"log"

	//"github.com/prometheus/client_golang/prometheus/promhttp"
	//"log"
	"log/slog"
	"net/http"
)

const namespace = "unisphere"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"unisphere_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"unisphere_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

type UnisphereCollectorSt struct {
	Collectors map[string]Collector
	Client     *client.UnisphereClient
}

var UnisphereCollector UnisphereCollectorSt

func Probe(w http.ResponseWriter, r *http.Request, logger *slog.Logger, reg *prometheus.Registry) {
	params := r.URL.Query()
	target := params.Get("target")
	module := params.Get("module")

	if target == "" {
		http.Error(w, "Target parameter missing or empty", http.StatusBadRequest)
		logger.Error("Target parameter missing or empty.")
		return
	} else if module == "" {
		http.Error(w, "Target parameter missing or empty", http.StatusBadRequest)
		logger.Error("Module parameter missing or empty.")
		return
	}

	uc, _ := client.NewClient(target, module, logger)
	//if !conected {
	//	return
	//}

	u := &UnisphereCollector
	u.Client = uc
	reg.MustRegister(u)
	log.Println()

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

}

type Collector interface {
	Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64
}

func NewCollector(cName string, c Collector) *UnisphereCollectorSt {
	u := &UnisphereCollector
	if u.Collectors == nil {
		u.Collectors = make(map[string]Collector)
	}
	u.Collectors[cName] = c

	return u
}

func (c *UnisphereCollectorSt) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}
func (c *UnisphereCollectorSt) Collect(ch chan<- prometheus.Metric) {
	for cName, collector := range c.Collectors {
		execute(cName, collector, c.Client, ch)
		log.Println(collector)
	}
}

func execute(cName string, c Collector, uc *client.UnisphereClient, ch chan<- prometheus.Metric) {
	start := time.Now()
	success := c.Update(uc, ch)
	duration := time.Since(start)
	if success == 0.0 {
		uc.Logger.Debug("Failed to Collect Metrics", "collector", cName)
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, 2, duration.Seconds(), cName)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, 2, success, cName)
}

func BuildFQName(sub string, name string) string {
	return prometheus.BuildFQName(namespace, sub, name)
}
