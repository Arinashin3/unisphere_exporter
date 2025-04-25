package collector

func init() {
	NewCollector(NewMetricLunCollector())
}

func NewMetricLunCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime_lun"
	m.metricPath = []string{
		"sp.*.storage.lun.*.sumOutstandingRequests",
		"sp.*.storage.lun.*.totalCallsRate",
		"sp.*.storage.lun.*.totalIoTime",
		"sp.*.storage.lun.*.writeBlocks",
		"sp.*.storage.lun.*.writeBytesRate",
		"sp.*.storage.lun.*.writes",
		"sp.*.storage.lun.*.writesRate",
		"sp.*.storage.lun.*.avgReadSize",
		"sp.*.storage.lun.*.avgWriteSize",
		"sp.*.storage.lun.*.busyTime",
		"sp.*.storage.lun.*.currentIOCount",
		"sp.*.storage.lun.*.idleTime",
		"sp.*.storage.lun.*.queueLength",
		"sp.*.storage.lun.*.readBlocks",
		"sp.*.storage.lun.*.readBytesRate",
		"sp.*.storage.lun.*.reads",
		"sp.*.storage.lun.*.readsRate",
		"sp.*.storage.lun.*.responseTime",
	}

	m.GenerateCollector()

	return m.subName, &m
}
