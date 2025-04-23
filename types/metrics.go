package types

type MetricPath struct {
	Paths    []string `json:"paths"`
	Interval int      `json:"interval"`
}

// MetricQueryEntries is response from querying a MetricCollection
type MetricQueryEntries struct {
	Entries []struct {
		Content struct {
			Id        int                    `json:"id"`
			Path      string                 `json:"path"`
			Timestamp string                 `json:"timestamp"`
			Values    map[string]interface{} `json:"values"`
		} `json:"content"`
	} `json:"entries"`
}
