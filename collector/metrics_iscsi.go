package collector

func init() {
	NewCollector(NewMetricIscsiCollector())
}

func NewMetricIscsiCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "realtime"
	m.metricPath = []string{
		"sp.*.iscsi.fePort.*.readBlocks",
		"sp.*.iscsi.fePort.*.readBytesRate",
		"sp.*.iscsi.fePort.*.reads",
		"sp.*.iscsi.fePort.*.readsRate",
		"sp.*.iscsi.fePort.*.writeBlocks",
		"sp.*.iscsi.fePort.*.writeBytesRate",
		"sp.*.iscsi.fePort.*.writes",
		"sp.*.iscsi.fePort.*.writesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}
