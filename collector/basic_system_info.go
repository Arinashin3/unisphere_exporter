package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

const basicSystemSubName = "basicsystem"

type BasicSystemCollector struct {
	subName string
}

func NewBasicSystemCollector() Collector {
	return &BasicSystemCollector{
		subName: basicSystemSubName,
	}
}

func init() {
	subName := "basicsystem"
	NewCollector(subName, NewBasicSystemCollector())
}

var (
	infoDesc = prometheus.NewDesc(
		BuildFQName(basicSystemSubName, "info"),
		"",
		[]string{"id", "model", "sw_ver", "api_ver"}, nil,
	)
)

func (c *BasicSystemCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.BasicSystemInfo
	path := "/api/types/basicSystemInfo/instances"
	resp := uc.Get(path, "compact=true")
	if resp == nil {
		return 0.0
	}
	err := json.Unmarshal(resp, &jData)
	if err != nil {
		uc.Logger.Error("Unmarshal Error", "path", path, "error_msg", err)
	}
	for _, content := range jData.Entries {
		d := content.Content
		ch <- prometheus.MustNewConstMetric(infoDesc, 2, 1.0, d.ID, d.Model, d.SoftwareVersion, d.ApiVersion)
	}
	return 1.0
}
