package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func init() {
	NewCollector(NewBasicSystemCollector())
}

type BasicSystemCollector struct {
	path     string
	infoDesc *prometheus.Desc
}

func NewBasicSystemCollector() (string, Collector) {
	subName := "basicsystem"
	path := "/api/types/basicSystemInfo/instances"
	labels := []string{"id", "model", "sw_ver", "api_ver"}
	return subName, &BasicSystemCollector{
		path:     path,
		infoDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subName, "info"), "System Version", labels, nil),
	}
}

func (c *BasicSystemCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.BasicSystemInfoEntries
	var result float64
	resp := uc.Get(c.path, "compact=true")
	if resp == nil {
		return result
	}
	if json.Unmarshal(resp, &jData) != nil {
		uc.Logger.Error("Unmarshalling Error", "path", c.path)
		return result
	}
	for _, content := range jData.Entries {
		d := content.Content
		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1.0, d.ID, d.Model, d.SoftwareVersion, d.ApiVersion)
	}
	result = 1.0
	return result
}
