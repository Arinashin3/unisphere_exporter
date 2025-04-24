package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"unisphere_exporter/types"
)

func (uc *UnisphereClient) PostMetricRealTimeQuery(metricPath []string, interval int64) int64 {
	uc.url.Path = "/api/types/metricRealTimeQuery/instances.json"
	var qid int64
	query := types.MetricRealTimeQueryRequest{
		Paths:    metricPath,
		Interval: interval,
	}
	qData, err := json.Marshal(query)
	if err != nil {
		uc.Logger.Error("Failed Marchalling", "error_msg", err)
		return qid
	}
	buff := bytes.NewBuffer(qData)
	req, err := http.NewRequest("POST", uc.url.String(), buff)
	if err != nil {
		uc.Logger.Error("Failed create NewRequest", "error_msg", err)
		return qid
	}
	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Error("Request Error", "error_msg", err)
		return qid
	}

	var jData types.MetricRealTimeQueryResponse
	body, err := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &jData) != nil {
		uc.Logger.Error("Unmarshalling Error", "path", uc.url.Path)
		return qid
	}
	qid = jData.Content.Id
	return qid

}

func (uc *UnisphereClient) GetMetricRealTimeQueryResult(qid int64) []byte {
	uc.url.Path = "/api/types/metricQueryResult/instances.json"
	uc.url.RawQuery = "queryId EQ " + strconv.FormatInt(qid, 10)
	req, err := http.NewRequest("GET", uc.url.String(), nil)
	if err != nil {
		uc.Logger.Error("Failed create NewRequest", "error_msg", err)
		return nil
	}
	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Error("Request Error", "error_msg", err)
		return nil
	}

	var jData types.MetricRealTimeQueryResult
	body, err := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &jData) != nil {
		uc.Logger.Error("Unmarshalling Error", "path", uc.url.Path)
		return nil
	}
	qid = jData
	return qid

}
