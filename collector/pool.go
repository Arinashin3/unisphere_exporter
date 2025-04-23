package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func init() {
	NewCollector(NewPoolCollector())
}

type PoolCollector struct {
	path          string
	raidTypeDesc  *prometheus.Desc
	sizeFreeDesc  *prometheus.Desc
	sizeTotalDesc *prometheus.Desc
	sizeUsedDesc  *prometheus.Desc
}

func NewPoolCollector() (string, Collector) {
	subName := "pool"
	path := "/api/types/pool/instances.json"
	labels := []string{"id", "name"}

	return subName, &PoolCollector{
		path: path,
		raidTypeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "raidtype"),
			"RAID group types or RAID levels. | 0:none, 1:raid5, 2:raid0, 3:raid1, 4:raid3, 7:raid10 ,10:raid6 ,12:mixed 48879:automatic.",
			labels, nil,
		),
		sizeFreeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_free"),
			"Size of free space available in the pool.",
			labels, nil,
		),
		sizeTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_total"),
			"The total size of space from the pool, which will be the sum of sizeFree, sizeUsed and size Preallocated space.",
			labels, nil,
		),
		sizeUsedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_used"),
			"Space allocated from the pool by storage resources, used for storing data. This will be the sum of the sizeAllocated values of each storage resource in the pool.",
			labels, nil,
		),
	}
}

func (c *PoolCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.PoolEntries
	var result float64
	resp := uc.Get(c.path, "compact=true")
	if resp == nil {
		return result
	}
	if json.Unmarshal(resp, &jData) != nil {
		uc.Logger.Error("Unmarshalling Error", "path", c.path)
		return result
	}
	for _, entries := range jData.Entries {
		d := entries.Content
		ch <- prometheus.MustNewConstMetric(c.raidTypeDesc, prometheus.GaugeValue, float64(d.RaidType), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.sizeFreeDesc, prometheus.GaugeValue, float64(d.SizeFree), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.sizeTotalDesc, prometheus.GaugeValue, float64(d.SizeTotal), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.sizeUsedDesc, prometheus.GaugeValue, float64(d.SizeUsed), d.ID, d.Name)
	}

	result = 1.0
	return result
}
