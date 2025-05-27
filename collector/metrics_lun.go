package collector

func init() {
	NewCollector(NewMetricLunCollector())
}

func NewMetricLunCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "lun"
	m.metricPath = []string{
		"sp.*.storage.lun.*.readBytesRate",
		"sp.*.storage.lun.*.readsRate",
		"sp.*.storage.lun.*.writeBytesRate",
		"sp.*.storage.lun.*.writesRate",
		"sp.*.storage.lun.*.queueLength",
		"sp.*.storage.lun.*.responseTime",
	}

	m.GenerateCollector()

	return m.subName, &m
}
