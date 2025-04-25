package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

type MetricSet struct {
	metricDesc *prometheus.Desc
	labelSet   map[string]string
}

type MetricCollector struct {
	subName    string
	metricPath []string
	metricList map[string]MetricSet
	logger     *slog.Logger
}

func (c *MetricCollector) GenerateCollector() {
	mList := make(map[string]MetricSet)
	for _, mPath := range c.metricPath {
		mList[mPath] = makeDesc(c.subName, mPath)
	}
	c.metricList = mList
}

func makeDesc(subName string, mPath string) MetricSet {
	name := strings.ToLower(mPath)
	labelSet := make(map[string]string)

	var labelList []string
	var before string
	var lastname string
	arr := strings.Split(name, ".")
	for i, v := range arr {
		now := v
		if now == "*" {
			if !(before == "") && !(before == "*") {
				labelList = append(labelList, before)
				labelSet[before] = ""
			}
		} else {
			if !(before == "") && !(before == "*") {
				if lastname == "" {
					lastname = before
				} else {
					lastname = lastname + "_" + before
				}
			}
		}
		before = now
		if i == len(arr)-1 {
			lastname = lastname + "_" + now
		}
	}
	fqName := prometheus.BuildFQName(namespace, subName, lastname)
	metricDesc := prometheus.NewDesc(fqName, "metric Path by - "+mPath, labelList, nil)

	return MetricSet{
		metricDesc: metricDesc,
		labelSet:   labelSet,
	}
}

func (c *MetricCollector) GenerateEntries(entries *types.MetricQueryEntries, ch chan<- prometheus.Metric) float64 {
	var result float64
	var f float64
	var push bool
	var err error
	for _, entry := range entries.Entries {
		content := entry.Content
		// Check to exist metricList
		if c.metricList[content.Path].metricDesc == nil {
			c.logger.Warn("cannot search the metric desc", "metric_path", content.Path)
			continue
		}
		for k1, v1 := range content.Values {
			v1type := reflect.TypeOf(v1).String()
			if strings.Contains(v1type, "interface") {
				for k2, v2 := range v1.(map[string]interface{}) {
					push = true
					v2type := reflect.TypeOf(v2).String()
					if strings.Contains(v2type, "interface") {
						c.logger.Error("parsing error - interface")
						push = false
					} else if strings.Contains(v2type, "int") {
						f = float64(reflect.ValueOf(v2).Int())
					} else if strings.Contains(v2type, "float") {
						f = reflect.ValueOf(v2).Float()
					} else if strings.Contains(v2type, "string") {
						f, err = strconv.ParseFloat(reflect.ValueOf(v2).String(), 64)
						if err != nil {
							c.logger.Error("parsing error - string")
							push = false
						}
					}
					if push {
						ch <- prometheus.MustNewConstMetric(c.metricList[content.Path].metricDesc, prometheus.GaugeValue, f, k1, k2)
					}
				}
			} else {
				push = true
				if strings.Contains(v1type, "int") {
					f = float64(reflect.ValueOf(v1).Int())
				} else if strings.Contains(v1type, "float") {
					f = reflect.ValueOf(v1).Float()
				} else if strings.Contains(v1type, "string") {
					f, err = strconv.ParseFloat(reflect.ValueOf(v1).String(), 64)
					if err != nil {
						c.logger.Error("parsing error - string")
						push = false
					}
				}
				if push {
					ch <- prometheus.MustNewConstMetric(c.metricList[content.Path].metricDesc, prometheus.GaugeValue, f, k1)
				}
			}
		}
	}
	return result
}

func (c *MetricCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var result float64
	var entries types.MetricQueryEntries
	c.logger = uc.Logger

	qid := uc.PostMetricRealTimeQuery(c.metricPath, 60)
	if qid == 0 {
		return result
	}

	data := uc.GetMetricRealTimeQueryResult(qid)
	err := json.Unmarshal(data, &entries)
	if err != nil {
		uc.Logger.Error("Unmarshalling Error", "error_msg", err)
	}
	result = c.GenerateEntries(&entries, ch)
	return result
}
