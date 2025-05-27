package collector

func init() {
	NewCollector(NewMetricFCCollector())
}

func NewMetricFCCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "fc"
	m.metricPath = []string{
		"sp.*.fibreChannel.fePort.*.readBytesRate",
		"sp.*.fibreChannel.fePort.*.readsRate",
		"sp.*.fibreChannel.fePort.*.writeBytesRate",
		"sp.*.fibreChannel.fePort.*.writesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}
