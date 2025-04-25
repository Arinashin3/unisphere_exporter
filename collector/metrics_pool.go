package collector

func init() {
	NewCollector(NewMetricPoolCollector())
}

func NewMetricPoolCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime_pool"
	m.metricPath = []string{
		"sp.*.storage.pool.*.dataSizeSubscribed",
		"sp.*.storage.pool.*.dataSizeUsed",
		"sp.*.storage.pool.*.overheadSizeSubscribed",
		"sp.*.storage.pool.*.overheadSizeUsed",
		"sp.*.storage.pool.*.sizeFree",
		"sp.*.storage.pool.*.sizeSubscribed",
		"sp.*.storage.pool.*.sizeTotal",
		"sp.*.storage.pool.*.sizeUsed",
		"sp.*.storage.pool.*.sizeUsedBlocks",
		"sp.*.storage.pool.*.snapshotSizeSubscribed",
		"sp.*.storage.pool.*.snapshotSizeUsed",
	}

	m.GenerateCollector()

	return m.subName, &m
}
