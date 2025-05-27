package collector

func init() {
	NewCollector(NewMetricGlobalCollector())
}

func NewMetricGlobalCollector() (string, Collector) {
	var m MetricCollector
	m.subName = "global"
	m.metricPath = []string{
		"sp.*.cpu.summary.utilization",
		"sp.*.blockCache.global.summary.dirtyBytes",
		"sp.*.net.basic.inBytesRate",
		"sp.*.net.basic.outBytesRate",
	}

	m.GenerateCollector()

	return m.subName, &m
}
