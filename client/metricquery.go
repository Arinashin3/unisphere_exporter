package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"unisphere_exporter/types"
)

func (uc *UnisphereClient) PostMetricRealTimeQuery(metricPath []string, interval int64) int64 {
	tgt := uc.url
	tgt.Path = "/api/types/metricRealTimeQuery/instances"
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
	req, err := http.NewRequest("POST", tgt.String(), buff)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.Header.Add("EMC-CSRF-TOKEN", uc.token)
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

func (uc *UnisphereClient) QueryMetricValue(path string) []byte {
	tgt := uc.url
	tgt.Path = "/api/types/metricValue/instances"
	tgt.RawQuery = "filter=path%20EQ%20" + "\"" + path + "\""
	req, err := http.NewRequest("GET", tgt.String(), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	if err != nil {
		uc.Logger.Error("Failed create NewRequest", "error_msg", err)
		return nil
	}
	resp, err := uc.hc.Do(req)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.Logger.Error("Request Error", "error_msg", err)
		return nil
	}
	if resp.StatusCode == 400 {
		uc.Logger.Error("Request Error", "error_msg", err)
		return nil
	}

	if err != nil {
		uc.Logger.Error("Unmarshalling Error", "error_msg", err)
		return nil
	}
	return body
}

func (uc *UnisphereClient) GetMetricRealTimeQueryResult(qid int64) []byte {
	tgt := uc.url
	tgt.Path = "/api/types/metricQueryResult/instances"
	tgt.RawQuery = "filter=path%20EQ%20"
	req, err := http.NewRequest("GET", tgt.String(), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+uc.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	if err != nil {
		uc.Logger.Error("Failed create NewRequest", "error_msg", err)
		return nil
	}
	resp, err := uc.hc.Do(req)
	if err != nil {
		uc.Logger.Error("Request Error", "error_msg", err)
		return nil
	}
	if resp.StatusCode == 400 {
		uc.Logger.Error("Request Error", "error_msg", err)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.Logger.Error("Unmarshalling Error", "error_msg", err)
		return nil
	}
	return body
}
