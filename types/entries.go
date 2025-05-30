package types

// Basic System Information Entries
// Path : "/api/types/basicSystemInfo/instances"
type BasicSystemInfoEntries struct {
	Entries []struct {
		Content BasicSystemInfoContent `json:"content"`
	} `json:"entries"`
}

type BasicSystemInfoContent struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Model               string `json:"model"`
	SoftwareVersion     string `json:"softwareVersion"`
	SoftwareFullVersion string `json:"softwareFullVersion"`
	ApiVersion          string `json:"apiVersion"`
	EarliestApiVersion  string `json:"earliestApiVersion"`
}

// Pool Entries
// Path : "/api/types/pool/instances"
type PoolEntries struct {
	Entries []struct {
		Content PoolContent `json:"content"`
	} `json:"entries"`
}

type PoolContent struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	RaidType  int    `json:"raidType"`
	SizeFree  uint64 `json:"sizeFree"`
	SizeTotal uint64 `json:"sizeTotal"`
}

// Lun Entries
// Path : "/api/types/lun/instances"
type LunEntries struct {
	Entries []struct {
		Content LunContent `json:"content"`
	} `json:"entries"`
}

type LunContent struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	SizeTotal     uint64 `json:"sizeTotal"`
	SizeAllocated uint64 `json:"sizeAllocated"`
	Pool          struct {
		Id string `json:"id,omitempty"`
	} `json:"pool"`
}

type SysCapEntries struct {
	Entries []struct {
		Content SysCapContent `json:"content"`
	} `json:"entries"`
}

type SysCapContent struct {
	Id                     string  `json:"id"`
	SizeFree               uint64  `json:"sizeFree"`
	SizeTotal              uint64  `json:"sizeTotal"`
	SizeUsed               uint64  `json:"sizeUsed"`
	SizePreallocated       uint64  `json:"sizePreallocated"`
	SizeSubscribed         uint64  `json:"sizeSubscribed"`
	DataReductionSizeSaved uint64  `json:"dataReductionSizeSaved"`
	DataReductionPercent   uint64  `json:"dataReductionPercent"`
	DataReductionRatio     float64 `json:"dataReductionRatio"`
	TotalLogicalSize       uint64  `json:"totalLogicalSize"`
	//ThinSavingRatio        float64 `json:"thinSavingRatio"`
	SnapsSavingsRatio      float64 `json:"snapsSavingsRatio"`
	OverallEfficiencyRatio float64 `json:"overallEfficiencyRatio"`
}

type FilesystemEntries struct {
	Entries []struct {
		Content FilesystemContent `json:"content"`
	} `json:"entries"`
}

type FilesystemContent struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	SizeTotal     uint64 `json:"sizeTotal"`
	SizeUsed      uint64 `json:"sizeUsed"`
	SizeAllocated uint64 `json:"sizeAllocated"`
	IsThinEnabled bool   `json:"isThinEnabled"`
}

type FcPortEntries struct {
	Entries []struct {
		Content FilesystemContent `json:"content"`
	} `json:"entries"`
}

type FcPortContent struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	CurrentSpeed uint64 `json:"sizeTotal"`
}
