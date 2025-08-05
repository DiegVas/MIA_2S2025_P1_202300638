package structs

type MBR struct {
	Mbr_size int64  `json:"mbr_size"`
	Mbr_timestamp [32]byte `json:"mbr_timestamp"`
	Mbr_uuid [36]byte `json:"mbr_uid"`
	Mbr_partitions int `json:"mbr_partitions"`
	Dsk_fit byte `json:"dsk_fit"`

}

type PARTITION struct {
	Part_status byte  `json:"part_status"`
	Part_type byte    `json:"part_type"`
	Part_fit byte     `json:"part_fit"`
	Part_start int64   `json:"part_start"`
	Part_size int64    `json:"part_size"`
	Part_name string   `json:"part_name"`
	Part_uuid string   `json:"part_uuid"`
	Part_correlative int    `json:"part_correlative"` 
}

type EBR struct {
	Part_mount byte   `json:"part_mount"`
	Part_fit byte     `json:"part_fit"`
	Part_start int64   `json:"part_start"`
	Part_size int64    `json:"part_size"`
	Part_next int64    `json:"part_next"`
	Part_name string   `json:"part_name"`
}