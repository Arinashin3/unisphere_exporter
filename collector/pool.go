package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

const poolSubName = "pool"

type PoolCollector struct {
	subName string
}

func init() {
	NewCollector(poolSubName, NewPoolCollector())
}

func NewPoolCollector() Collector {
	return &PoolCollector{subName: poolSubName}
}

var (
	raidTypeDesc = prometheus.NewDesc(
		BuildFQName(poolSubName, "raidtype"),
		"RAID group types or RAID levels. | 0:none, 1:raid5, 2:raid0, 3:raid1, 4:raid3, 7:raid10 ,10:raid6 ,12:mixed 48879:automatic.",
		[]string{"id", "name"}, nil,
	)
	sizeFreeDesc = prometheus.NewDesc(
		BuildFQName(poolSubName, "size_free"),
		"Size of free space available in the pool.",
		[]string{"id", "name"}, nil,
	)
	sizeTotalDesc = prometheus.NewDesc(
		BuildFQName(poolSubName, "size_total"),
		"The total size of space from the pool, which will be the sum of sizeFree, sizeUsed and size Preallocated space.",
		[]string{"id", "name"}, nil,
	)
	sizeUsedDesc = prometheus.NewDesc(
		BuildFQName(poolSubName, "size_used"),
		"Space allocated from the pool by storage resources, used for storing data. This will be the sum of the sizeAllocated values of each storage resource in the pool.",
		[]string{"id", "name"}, nil,
	)
)

func (c *PoolCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.PoolInfo
	path := "/api/types/pool/instances"
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
		ch <- prometheus.MustNewConstMetric(raidTypeDesc, 2, float64(d.RaidType), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(sizeFreeDesc, 2, float64(d.SizeFree), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(sizeTotalDesc, 2, float64(d.SizeTotal), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(sizeUsedDesc, 2, float64(d.SizeUsed), d.ID, d.Name)

	}

	return 1.0
}
