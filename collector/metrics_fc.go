package collector

func init() {
	NewCollector(NewMetricFCCollector())
}

func NewMetricFCCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime_fc"
	m.metricPath = []string{
		"sp.*.fibreChannel.fePort.*.readBlocks",
		"sp.*.fibreChannel.fePort.*.readBytesRate",
		"sp.*.fibreChannel.fePort.*.reads",
		"sp.*.fibreChannel.fePort.*.readsRate",
		"sp.*.fibreChannel.fePort.*.writeBlocks",
		"sp.*.fibreChannel.fePort.*.writeBytesRate",
		"sp.*.fibreChannel.fePort.*.writes",
		"sp.*.fibreChannel.fePort.*.writesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}
