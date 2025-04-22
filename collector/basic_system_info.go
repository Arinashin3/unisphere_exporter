package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
)

//type BasicSystemInfoEntries struct {
//	Entries []types.BasicSystemInfo `json:"entries"`
//}
//
//var (
//	info = prometheus.NewDesc("test", "help", nil, nil)
//)
//
//type BasicSystemInfoDesc struct {
//	info *prometheus.Collector
//	ver  *prometheus.Collector
//}

type BasicSystemCollector struct {
	subName string
}

type BasicSystemInfo struct {
	Entries []struct {
		Content struct {
			ID                 string `json:"id"`
			Model              string `json:"model"`
			Name               string `json:"name"`
			SoftwareVersion    string `json:"softwareVersion"`
			APIVersion         string `json:"apiVersion"`
			EarliestAPIVersion string `json:"earliestApiVersion"`
		} `json:"content"`
	} `json:"entries"`
}

func NewBasicSystemCollector(sub string) Collector {
	return &BasicSystemCollector{
		subName: sub,
	}
}

func init() {
	subName := "basicsystem"
	NewCollector(subName, NewBasicSystemCollector(subName))
}

func (c *BasicSystemCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	//var jData BasicSystemInfo
	resp := uc.Get("/api/types/basicSystemInfo/instances", "compact=true")
	if resp == nil {
		return 0.0
	}
	//json.Unmarshal(resp, &jData)

	return 1.0
}

//type BasicSystemDesc struct {
//	infoDesc *prometheus.Desc
//}
//
//
//type BasicSystemCollector struct {
//	subsystem string
//	metricDesc *prometheus.Desc
//}
//func NewBasicSystemCollector() Collector {
//
//return &BasicSystemCollector{subsystem: namespace}
//
//}
//
//type BasicSystemCollector struct {
//	info *typedDesc
//}

//func (c *BasicSystemCollector) Update(ch chan<- prometheus.Metric) float64 {
//
//}

//func getBasicSystemInfo(uc *client.UnisphereClient, reg *prometheus.Registry) (BasicSystemInfoEntries, bool) {
//
//	var resp BasicSystemInfoEntries
//	err := c.Get("/api/types/basicSystemInfo/instances", "compact=true", &resp)
//	if err != nil {
//		log.Printf("Error getting basic system info: %s", err)
//		return BasicSystemInfoEntries{}, false
//	}
//
//	return resp, true
//
//}
//func ProbeBasicSystemInfo(uc client.UnisphereClient, registry *prometheus.Registry) bool {
//	// Variable qr is return value
//	qr := false
//
//	//data, qr := getBasicSystemInfo(uc)
//	if qr == false {
//		return qr
//	}
//	labels := []string{"system_name"}
//
//	mBasicSysInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "unisphere_basic_system_info", Help: "This storage systems's Infomation"}, append(labels, "model"))
//	mBasicSysSoftVersion := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "unisphere_basic_system_software_version", Help: "Software version of this storage system."}, append(labels, "version", "full_version"))
//	mBasicSysApiVersion := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "unisphere_basic_system_api_version", Help: "Latest REST API Version, that this storage system supports."}, append(labels, "api_version"))
//	mBasicSysEarliestApiVersion := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "unisphere_basic_system_api_version", Help: "Earliest REST API Version, that this storage system supports."}, append(labels, "api_version"))
//
//	registry.MustRegister(mBasicSysInfo)
//	registry.MustRegister(mBasicSysSoftVersion)
//	registry.MustRegister(mBasicSysApiVersion)
//	registry.MustRegister(mBasicSysEarliestApiVersion)
//
//	for _, entry := range data.Entries {
//		mBasicSysInfo.WithLabelValues(entry.Content.Name, entry.Content.Model).Set(1)
//		mBasicSysSoftVersion.WithLabelValues(entry.Content.Name, entry.Content.SoftwareVersion, entry.Content.SoftwareFullVersion).Set(1)
//		mBasicSysApiVersion.WithLabelValues(entry.Content.Name, entry.Content.ApiVersion).Set(1)
//		mBasicSysEarliestApiVersion.WithLabelValues(entry.Content.Name, entry.Content.EarliestApiVersion).Set(1)
//	}
//
//	return true
//}
