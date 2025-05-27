package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"unisphere_exporter/client"
	"unisphere_exporter/types"
)

func init() {
	NewCollector(NewPoolCollector())
}

type PoolCollector struct {
	apiPath        string
	requiredFields []string
	compact        bool
	content        types.PoolContent
	descList       map[string]*prometheus.Desc
}

func NewPoolCollector() (string, Collector) {
	subName := "pool"
	labels := []string{"id", "name"}
	var c PoolCollector

	c.apiPath = "/api/types/pool/instances"
	c.descList = make(map[string]*prometheus.Desc)
	c.compact = true

	for _, field := range reflect.VisibleFields(reflect.TypeOf(c.content)) {
		fType := field.Type.String()
		fName := strings.Trim(string(field.Tag), "json:")
		fName = strings.Trim(fName, "\"")
		if fType != "string" {
			c.descList[field.Name] = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subName, fName),
				fName+" metric",
				labels, nil,
			)
		}
		c.requiredFields = append(c.requiredFields, fName)
	}

	return subName, &c
}

func (c *PoolCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.PoolEntries
	var result float64

	query := "fields=" + strings.Join(c.requiredFields, ",")
	if c.compact {
		query += "&compact=true"
	}
	resp := uc.Get(c.apiPath, query)
	if resp == nil {
		return result
	}
	if json.Unmarshal(resp, &jData) != nil {
		uc.Logger.Error("Unmarshalling Error", "path", c.apiPath)
		return result
	}
	for _, entries := range jData.Entries {
		d := entries.Content
		ch <- prometheus.MustNewConstMetric(c.descList["RaidType"], prometheus.GaugeValue, float64(d.RaidType), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.descList["SizeFree"], prometheus.GaugeValue, float64(d.SizeFree), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.descList["SizeTotal"], prometheus.GaugeValue, float64(d.SizeTotal), d.ID, d.Name)
		ch <- prometheus.MustNewConstMetric(c.descList["SizeUsed"], prometheus.GaugeValue, float64(d.SizeUsed), d.ID, d.Name)
	}

	result = 1.0
	return result
}
