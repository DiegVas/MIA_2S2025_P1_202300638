package structs

type PartitionInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Fit    string `json:"fit"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
}

type DiskPartitionsInfo struct {
	DiskPath   string          `json:"disk_path"`
	Partitions []PartitionInfo `json:"partitions"`
}
