package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sync"
	"unisphere_exporter/types"
	"unisphere_exporter/utils"
)

type SysCapacityEntries struct {
	Entries []types.SysCapacity `json:"entries"`
}

func getSystemCapacity(c *utils.UnisphereClient) (SysCapacityEntries, float64) {
	var resp SysCapacityEntries
	err := c.Get("/api/types/System Capacity/instances", "compact=true", &resp)
	if err != nil {
		log.Printf("Error getting storage System Capacity info: %s", err)
		return SysCapacityEntries{}, 0.0
	}

	return resp, 1.0

}

func ProbeSystemCapacity(c utils.UnisphereClient, registry *prometheus.Registry, wg *sync.WaitGroup) float64 {
	// Variable qr is return value
	defer wg.Done()
	qr := 0.0

	data, qr := getSystemCapacity(&c)
	if qr == 0.0 {
		return qr
	}

	mSysCapacitySizeFree := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_size_free", Help: "Size of free space available in the System Capacity."})
	mSysCapacitySizeTotal := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_size_total", Help: "The total size of space from the System Capacity, which will be the sum of sizeFree, sizeUsed and sizePreallocated space."})
	mSysCapacitySizeUsed := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_size_used", Help: "Space allocated from the System Capacity by storage resources, used for storing data. This will be the sum of the sizeAllocated values of each storage resource in the System Capacity."})
	mSysCapacitySizePreallocated := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_size_preallocated", Help: "Space reserved form the System Capacity by storage resources, for future needs to make writes more efficient. The System Capacity may be able to reclaim some of this if space is running low. This will be the sum of the sizePreallocated values of each storage resource in the System Capacity."})
	mSysCapacitySizeSubscribed := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_size_subscribed", Help: "Size of space requested by the storage resources allocated in the System Capacity for possible future allocations. If this value is greater than the total size of the System Capacity, the System Capacity is considered oversubscribed."})
	mSysCapacityDataReductionSizeSaved := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_data_reduction_size_save", Help: "Amount of space saved for the System Capacity by data reduction (includes savings from compression, deduplication and advanced deduplication)."})
	mSysCapacityDataReductionPercent := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_data_reduction_pct", Help: "Data reduction percentage is the percentage of the data that does not consume storage - the savings due to data reduction. For example, if 1 TB of data is stored in 250 GB, the data reduction percentage is 75%. 75% data reduction percentage is equivalent to a 4:1 data reduction ratio."})
	mSysCapacityDataReductionRatio := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_data_reduction_ratio", Help: "Data reduction ratio. The data reduction ratio is the ratio between the size of the data and the amount of storage actually consumed. For example, 1TB of data consuming 250GB would have a ration of 4:1. A 4:1 data reduction ratio is equivalent to a 75% data reduction percentage."})
	mSysCapacityTotalLogicalSize := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_total_logical_size", Help: "Total logical provisioned capacity of primary storage objects visible to hosts (as defined by the 'size' attribute for each primary storage objects), plus the total logical provisioned capacity of all Snapshots (that is, the total capacity that would be required if every snapshot were a fully provisioned copy instead). "})
	mSysCapacityThinSavingRatio := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_thin_saving_ratio", Help: "Storage efficiency ratio of thin provisioned primary storage resources on the system, which demonstrates the efficiency of thin provisioning. This is calculated as follows: (Total provisioned size of all primary storage resources) / (Total allocated size of all primary storage resources without including data reduction savings, if any). Because savings due to thin provisioning are virtual savings, this could exceed the total system capacity. This does not include data reduction savings or snapshot savings. "})
	mSysCapacitySnapsSavingRatio := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_snaps_saving_ratio", Help: "Storage efficiency ratio of snapshots on the system, calculated based on the capacity that would have been required for fully provisioned copies, which demonstrates the efficiency of snapshots. This is calculated as follows: (Total snapshot size of all snapshots as if they were full copies) / (Total allocated size of all snapshots without including data reduction savings, if any). This does not include data reduction savings or thin provisioning savings. "})
	mSysCapacityOverallEfficiencyRatio := prometheus.NewGauge(prometheus.GaugeOpts{Name: "unisphere_sys_capacity_overall_efficiency_ratio", Help: "System-level storage efficiency ratio, calculated by dividing the total logical capacity of the System by the actual Used capacity on the System, This leverages the efficiency features of thin provisioning, snapshots and data reduction(compression and deduplication). "})

	registry.MustRegister(mSysCapacitySizeFree)
	registry.MustRegister(mSysCapacitySizeTotal)
	registry.MustRegister(mSysCapacitySizeUsed)
	registry.MustRegister(mSysCapacitySizePreallocated)
	registry.MustRegister(mSysCapacitySizeSubscribed)
	registry.MustRegister(mSysCapacityDataReductionSizeSaved)
	registry.MustRegister(mSysCapacityDataReductionPercent)
	registry.MustRegister(mSysCapacityDataReductionRatio)
	registry.MustRegister(mSysCapacityTotalLogicalSize)
	registry.MustRegister(mSysCapacityThinSavingRatio)
	registry.MustRegister(mSysCapacitySnapsSavingRatio)
	registry.MustRegister(mSysCapacityOverallEfficiencyRatio)

	for _, entry := range data.Entries {
		mSysCapacitySizeFree.Set(float64(entry.Content.SizeFree))
		mSysCapacitySizeTotal.Set(float64(entry.Content.SizeTotal))
		mSysCapacitySizeUsed.Set(float64(entry.Content.SizeUsed))
		mSysCapacitySizePreallocated.Set(float64(entry.Content.SizePreallocated))
		mSysCapacitySizeSubscribed.Set(float64(entry.Content.SizeSubscribed))
		mSysCapacityDataReductionSizeSaved.Set(float64(entry.Content.DataReductionSizeSaved))
		mSysCapacityDataReductionPercent.Set(float64(entry.Content.DataReductionPercent))
		mSysCapacityDataReductionRatio.Set(entry.Content.DataReductionRatio)
		mSysCapacityTotalLogicalSize.Set(float64(entry.Content.TotalLogicalSize))
		mSysCapacityThinSavingRatio.Set(entry.Content.ThinSavingRatio)
		mSysCapacitySnapsSavingRatio.Set(entry.Content.SnapsSavingsRatio)
		mSysCapacityOverallEfficiencyRatio.Set(entry.Content.OverallEfficiencyRatio)
	}

	return 1.0
}
