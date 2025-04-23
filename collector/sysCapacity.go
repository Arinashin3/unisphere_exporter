package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func init() {
	NewCollector(NewSysCapCollector())
}

type SysCapCollector struct {
	path                       string
	sizeFreeDesc               *prometheus.Desc
	sizeTotalDesc              *prometheus.Desc
	sizeUsedDesc               *prometheus.Desc
	sizePreallocatedDesc       *prometheus.Desc
	dataReductionSizeSavedDesc *prometheus.Desc
	dataReductionPercentDesc   *prometheus.Desc
	dataReductionRatioDesc     *prometheus.Desc
	totalLogicalSizeDesc       *prometheus.Desc
	thinSavingRatioDesc        *prometheus.Desc
	snapsSavingsRatioDesc      *prometheus.Desc
	overallEfficiencyRatioDesc *prometheus.Desc
}

func NewSysCapCollector() (string, Collector) {
	subName := "syscap"
	path := "/api/types/systemCapacity/instances"
	labels := []string{"id", "name"}
	return subName, &SysCapCollector{
		path: path,
		sizeFreeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_free"),
			"Size of free space available in the System Capacity.",
			labels, nil,
		),
		sizeTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_total"),
			"The total size of space from the System Capacity, which will be the sum of sizeFree, sizeUsed and sizePreallocated space.",
			labels, nil,
		),
		sizeUsedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_used"),
			"Space allocated from the System Capacity by storage resources, used for storing data. This will be the sum of the sizeAllocated values of each storage resource in the System Capacity.",
			labels, nil,
		),
		sizePreallocatedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_preallocated"),
			"Space reserved form the System Capacity by storage resources, for future needs to make writes more efficient. The System Capacity may be able to reclaim some of this if space is running low. This will be the sum of the sizePreallocated values of each storage resource in the System Capacity.",
			labels, nil,
		),
		dataReductionSizeSavedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_preallocated"),
			"Amount of space saved for the System Capacity by data reduction.",
			labels, nil,
		),
		dataReductionPercentDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_preallocated"),
			"Data reduction percentage is the percentage of the data that does not consume storage - the savings due to data reduction.",
			labels, nil,
		),
		dataReductionRatioDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_preallocated"),
			"Data reduction ratio. The data reduction ratio is the ratio between the size of the data and the amount of storage actually consumed.",
			labels, nil,
		),
		totalLogicalSizeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "size_logicaltotal"),
			"Total logical provisioned capacity of primary storage objects visible to hosts, plus the total logical provisioned capacity of all Snapshots.",
			labels, nil,
		),
		thinSavingRatioDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "thinsaving_ratio"),
			"Storage efficiency ratio of thin provisioned primary storage resources on the system, which demonstrates the efficiency of thin provisioning.",
			labels, nil,
		),
		snapsSavingsRatioDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "snapsaving_ratio"),
			"Storage efficiency ratio of snapshots on the system, calculated based on the capacity that would have been required for fully provisioned copies, which demonstrates the efficiency of snapshots.",
			labels, nil,
		),
		overallEfficiencyRatioDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subName, "overallefficiency_ratio"),
			"System-level storage efficiency ratio, calculated by dividing the total logical capacity of the System by the actual Used capacity on the System, This leverages the efficiency features of thin provisioning, snapshots and data reduction(compression and deduplication)",
			labels, nil,
		),
	}
}

func (c *SysCapCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.SysCapEntries
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
		ch <- prometheus.MustNewConstMetric(c.sizeFreeDesc, prometheus.GaugeValue, float64(d.SizeFree), d.ID)
		ch <- prometheus.MustNewConstMetric(c.sizeTotalDesc, prometheus.GaugeValue, float64(d.SizeTotal), d.ID)
		ch <- prometheus.MustNewConstMetric(c.sizeUsedDesc, prometheus.GaugeValue, float64(d.SizeUsed), d.ID)
		ch <- prometheus.MustNewConstMetric(c.sizePreallocatedDesc, prometheus.GaugeValue, float64(d.SizePreallocated), d.ID)
		ch <- prometheus.MustNewConstMetric(c.dataReductionSizeSavedDesc, prometheus.GaugeValue, float64(d.DataReductionSizeSaved), d.ID)
		ch <- prometheus.MustNewConstMetric(c.dataReductionPercentDesc, prometheus.GaugeValue, float64(d.DataReductionPercent), d.ID)
		ch <- prometheus.MustNewConstMetric(c.dataReductionRatioDesc, prometheus.GaugeValue, float64(d.DataReductionRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.totalLogicalSizeDesc, prometheus.GaugeValue, float64(d.TotalLogicalSize), d.ID)
		ch <- prometheus.MustNewConstMetric(c.thinSavingRatioDesc, prometheus.GaugeValue, float64(d.ThinSavingRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.snapsSavingsRatioDesc, prometheus.GaugeValue, float64(d.SnapsSavingsRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.overallEfficiencyRatioDesc, prometheus.GaugeValue, float64(d.OverallEfficiencyRatio), d.ID)

	}

	result = 1.0
	return result
}
