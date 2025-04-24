package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type MetricSet struct {
	metricDesc *prometheus.Desc
	labelSet   map[string]string
}

type MetricCollector struct {
	subName    string
	metricPath []string
	metricList map[string]MetricSet
}

func (m *MetricCollector) GenerateCollector() {
	mList := make(map[string]MetricSet)
	for _, mPath := range m.metricPath {
		mList[mPath] = makeDesc(m.subName, mPath)
	}
	m.metricList = mList
}

func makeDesc(subName string, mPath string) MetricSet {
	name := strings.ToLower(mPath)
	sep := ".*."
	labelSet := make(map[string]string)

	var labelList []string
	var pre string
	for strings.Contains(name, sep) {
		pre, name, _ = strings.Cut(name, sep)
		arr := strings.Split(pre, ".")
		pre = strings.TrimLeft(arr[len(arr)-1], sep)
		labelList = append(labelList, pre)
		labelSet[pre] = ""
	}
	fqName := prometheus.BuildFQName(namespace, subName, name)
	metricDesc := prometheus.NewDesc(fqName, "metric Path by - "+mPath, labelList, nil)

	return MetricSet{
		metricDesc: metricDesc,
		labelSet:   labelSet,
	}
}
