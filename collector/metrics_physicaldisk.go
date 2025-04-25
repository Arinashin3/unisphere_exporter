package collector

func init() {
	NewCollector(NewMetricPhysicalDiskCollector())
}

func NewMetricPhysicalDiskCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime_disk"
	m.metricPath = []string{
		"sp.*.physical.disk.*.averageQueueLength",
		"sp.*.physical.disk.*.busyTicks",
		"sp.*.physical.disk.*.idleTicks",
		"sp.*.physical.disk.*.readBlocks",
		"sp.*.physical.disk.*.readBytesRate",
		"sp.*.physical.disk.*.reads",
		"sp.*.physical.disk.*.readsRate",
		"sp.*.physical.disk.*.responseTime",
		"sp.*.physical.disk.*.serviceTime",
		"sp.*.physical.disk.*.sumArrivalQueueLength",
		"sp.*.physical.disk.*.totalCallsRate",
		"sp.*.physical.disk.*.writeBlocks",
		"sp.*.physical.disk.*.writeBytesRate",
		"sp.*.physical.disk.*.writes",
		"sp.*.physical.disk.*.writesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}
