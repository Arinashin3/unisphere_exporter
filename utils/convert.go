package utils

import "github.com/tidwall/gjson"

type Bytes int64

func (b Bytes) ToKiB() float64 {
	return float64(b / 1024)
}

func (b Bytes) ToMiB() float64 {
	return float64(b / 1024 / 1024)
}

func (b Bytes) ToGiB() float64 {
	return float64(b / 1024 / 1024 / 1024)
}

func (b Bytes) ToTiB() float64 {
	return float64(b / 1024 / 1024 / 1024 / 1024)
}

func (b Bytes) ToPiB() float64 {
	return float64(b / 1024 / 1024 / 1024 / 1024 / 1024)
}

type Metric struct {
	Labels []string
	Value  gjson.Result
}

func ParseMetric(data gjson.Result) []*Metric {
	result, _, _ := parseMetric(nil, nil, data)
	return result
}

func parseMetric(result []*Metric, labels []string, data gjson.Result) ([]*Metric, []string, gjson.Result) {
	if data.IsObject() {
		for k, v := range data.Map() {
			tmpLabels := append(labels, k)
			result, _, _ = parseMetric(result, tmpLabels, v)
		}
	} else {
		var tmp = &Metric{}
		tmp.Labels = labels
		tmp.Value = data
		result = append(result, tmp)
	}
	return result, nil, gjson.Result{}
}
