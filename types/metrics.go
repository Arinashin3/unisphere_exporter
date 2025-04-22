package types

type MetricPath struct {
	Paths    []string `json:"paths"`
	Interval int      `json:"interval"`
}

// MetricResult is part of response of a MetricCollection query
type MetricResult struct {
	Id        int                    `json:"id"`
	Path      string                 `json:"path"`
	Timestamp string                 `json:"timestamp"`
	Values    map[string]interface{} `json:"values"`
}

// MetricResultEntry is part of response of a MetricCollection query
type MetricResultEntry struct {
	Content MetricResult `json:"content"`
}

// Pool
type PoolInfo struct {
	Entries []struct {
		Content struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			RaidType  int    `json:"raidType"`
			SizeFree  uint64 `json:"sizeFree"`
			SizeTotal uint64 `json:"sizeTotal"`
			SizeUsed  uint64 `json:"sizeUsed"`
		} `json:"content"`
	} `json:"entries"`
}

// MetricQueryResult is response from querying a MetricCollection
type MetricQueryResult struct {
	Base    string              `json:"base"`
	Updated string              `json:"updated"`
	Entries []MetricResultEntry `json:"entries"`
}

// StoragePool Struct to capture the response of StoragePool response
type StoragePool struct {
	Content StoragePoolContent `json:"content"`
}

// StoragePoolContent Struct to capture the StoragePool Content properties
type StoragePoolContent struct {
	ID                          string  `json:"id"`
	Name                        string  `json:"name"`
	RaidType                    int     `json:"raidType"`
	SizeFree                    uint64  `json:"sizeFree"`
	SizeTotal                   uint64  `json:"sizeTotal"`
	SizeUsed                    uint64  `json:"sizeUsed"`
	SizePreallocated            uint64  `json:"sizePreallocated"`
	SizeSubscribed              uint64  `json:"sizeSubscribed"`
	DataReductionSizeSaved      uint64  `json:"dataReductionSizeSaved"`
	DataReductionPercent        uint64  `json:"dataReductionPercent"`
	DataReductionRatio          float64 `json:"dataReductionRatio"`
	HasDataReductionEnabledLuns bool    `json:"hasDataReductionEnabledLuns"`
	HasDataReductionEnabledFs   bool    `json:"hasDataReductionEnabledFs"`
	IsFASTCacheEnabled          bool    `json:"isFASTCacheEnabled"`
	IsAllFlash                  bool    `json:"isAllFlash"`
}

// Basic System Information Entries
type BasicSystemInfo struct {
	Entries []struct {
		Content struct {
			ID                  string `json:"id"`
			Name                string `json:"name"`
			Model               string `json:"model"`
			SoftwareVersion     string `json:"softwareVersion"`
			SoftwareFullVersion string `json:"softwareFullVersion"`
			ApiVersion          string `json:"apiVersion"`
			EarliestApiVersion  string `json:"earliestApiVersion"`
		} `json:"content"`
	} `json:"entries"`
}

type SysCapacity struct {
	Content SysCapacityContent `json:"content"`
}

type SysCapacityContent struct {
	ID                     string  `json:"id"`
	SizeFree               uint64  `json:"sizeFree"`
	SizeTotal              uint64  `json:"sizeTotal"`
	SizeUsed               uint64  `json:"sizeUsed"`
	SizePreallocated       uint64  `json:"sizePreallocated"`
	SizeSubscribed         uint64  `json:"sizeSubscribed"`
	DataReductionSizeSaved uint64  `json:"dataReductionSizeSaved"`
	DataReductionPercent   uint64  `json:"dataReductionPercent"`
	DataReductionRatio     float64 `json:"dataReductionRatio"`
	TotalLogicalSize       uint64  `json:"totalLogicalSize"`
	ThinSavingRatio        float64 `json:"thinSavingRatio"`
	SnapsSavingsRatio      float64 `json:"snapsSavingsRatio"`
	OverallEfficiencyRatio float64 `json:"overallEfficiencyRatio"`
}
