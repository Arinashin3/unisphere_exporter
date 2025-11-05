package main

import (
	"fmt"
	"testing"

	"github.com/tidwall/gjson"
)

func TestName(t *testing.T) {
	testData := `{
		"id" : "Host_2",
		"health" : {
		"value" : 5
	},
		"name" : "VIOS02",
		"osType" : "AIX",
		"fcHostInitiators" : [ {
	"id" : "HostInitiator_22",
	"nodeWWN" : "20:00:00:90:FA:DC:66:B0",
	"portWWN" : "10:00:00:90:FA:DC:66:B0",
	"paths" : [ {
	"id" : "HostInitiator_22_02:00:01:07_0",
	"fcPort" : {
	"id" : "spb_iom_1_fc3"
	}
	} ]
	}, {
	"id" : "HostInitiator_30",
	"nodeWWN" : "20:00:00:90:FA:DC:66:B1",
	"portWWN" : "10:00:00:90:FA:DC:66:B1",
	"paths" : [ {
	"id" : "HostInitiator_30_02:00:00:07_0",
	"fcPort" : {
	"id" : "spa_iom_1_fc3"
	}
	} ]
	} ]
	}
	wwn := "50:06:01:60:CC:E0:48:9D:50:06:01:6F:4C:E0:48:9D"
	wwnn := wwn[:23]
	wwpn := wwn[24:]
	fmt.Println(wwnn, wwpn)

	tm := gjson.Get("{name: true}", "name").Float()
	td := gjson.Get("{name: false}", "name").Float()
	fmt.Println(tm, td)
	testData := "{\"queryId\":6,\"path\":\"sp.*.physical.coreCount\",\"timestamp\":\"2025-10-28T08:10:00.000Z\",\"values\":{ \"spa\": {\"data\": \"16\"}, \"spb\": \"16\"} }"

	data := gjson.Parse(testData)
	value := data.Get("values")
	// 넣을 데이터 : gjson.Result

	d, _, m := TMap([]*MetricSt{}, nil, value)

	for _, v := range value.Map() {
		m := v.IsObject()
		fmt.Println(m)

	}`
	data := gjson.Parse(testData)
	id := data.Get("id").String()
	name := data.Get("name").String()
	os := data.Get("osType").String()

	fmt.Println(id, name, os)

}

type MetricSt struct {
	Labels []string     `json:"labels"`
	Value  gjson.Result `json:"value"`
	Fin    bool         `json:"fin"`
}

func TMap(result []*MetricSt, labels []string, data gjson.Result) ([]*MetricSt, []string, gjson.Result) {
	if data.IsObject() {
		for k, v := range data.Map() {
			tmpLabels := append(labels, k)
			result, _, _ = TMap(result, tmpLabels, v)
		}
	} else {
		var tmp = &MetricSt{}
		tmp.Labels = labels
		tmp.Value = data
		result = append(result, tmp)
	}
	return result, nil, gjson.Result{}
}
