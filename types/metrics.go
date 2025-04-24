package types

type MetricRealTimeQueryRequest struct {
	Paths    []string `json:"paths"`
	Interval int64    `json:"interval"`
}

type MetricRealTimeQueryResponse struct {
	Content struct {
		Id int64 `json:"id"`
	} `json:"content"`
}

// MetricQueryEntries is response from querying a MetricCollection
type MetricQueryEntries struct {
	Entries []struct {
		Content struct {
			QueryId   int64                  `json:"queryId"`
			Path      string                 `json:"path"`
			Timestamp string                 `json:"timestamp"`
			Values    map[string]interface{} `json:"values"`
		} `json:"content"`
	} `json:"entries"`
}
