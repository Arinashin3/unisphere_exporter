package types

// Basic System Information Entries
// Path : "/api/types/basicSystemInfo/instances"
type BasicSystemInfoEntries struct {
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

// Pool Entries
// Path : "/api/types/pool/instances"
type PoolEntries struct {
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

type SysCapEntries struct {
	Entries []struct {
		Content struct {
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
		} `json:"content"`
	} `json:"entries"`
}
