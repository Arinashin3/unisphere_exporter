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
	NewCollector(NewSysCapCollector())
}

type SysCapCollector struct {
	apiPath        string
	requiredFields []string
	compact        bool
	content        types.SysCapContent
	descList       map[string]*prometheus.Desc
}

func NewSysCapCollector() (string, Collector) {
	subName := "syscap"
	labels := []string{"id"}
	var c SysCapCollector

	c.apiPath = "/api/types/systemCapacity/instances"
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

func (c *SysCapCollector) Update(uc *client.UnisphereClient, ch chan<- prometheus.Metric) float64 {
	var jData types.SysCapEntries
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
	if jData.Entries == nil {
		return result
	}
	for _, content := range jData.Entries {
		d := content.Content
		ch <- prometheus.MustNewConstMetric(c.descList["SizeFree"], prometheus.GaugeValue, float64(d.SizeFree), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["SizeTotal"], prometheus.GaugeValue, float64(d.SizeTotal), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["SizeUsed"], prometheus.GaugeValue, float64(d.SizeUsed), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["SizePreallocated"], prometheus.GaugeValue, float64(d.SizePreallocated), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["DataReductionSizeSaved"], prometheus.GaugeValue, float64(d.DataReductionSizeSaved), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["DataReductionPercent"], prometheus.GaugeValue, float64(d.DataReductionPercent), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["DataReductionRatio"], prometheus.GaugeValue, float64(d.DataReductionRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["TotalLogicalSize"], prometheus.GaugeValue, float64(d.TotalLogicalSize), d.ID)
		//ch <- prometheus.MustNewConstMetric(c.descList["ThinSavingRatio"], prometheus.GaugeValue, float64(d.ThinSavingRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["SnapsSavingsRatio"], prometheus.GaugeValue, float64(d.SnapsSavingsRatio), d.ID)
		ch <- prometheus.MustNewConstMetric(c.descList["OverallEfficiencyRatio"], prometheus.GaugeValue, float64(d.OverallEfficiencyRatio), d.ID)

	}

	result = 1.0
	return result
}
