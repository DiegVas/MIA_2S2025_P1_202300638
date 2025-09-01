package diskmanager

import (
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func readLogicalPartitions(diskPath string, extendedStart int32) ([]structs.PartitionInfo, error) {
	var logicalPartitions []structs.PartitionInfo

	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	currentPos := extendedStart
	for {
		ebrBytes := make([]byte, 30)
		_, err := file.ReadAt(ebrBytes, int64(currentPos))
		if err != nil {
			break
		}

		// Leer los campos del EBR usando la estructura correcta
		// Byte 0: Part_mount
		// Byte 1: Part_fit
		// Bytes 2-5: Part_start (int32)
		// Bytes 6-9: Part_size (int32)
		// Bytes 10-13: Part_next (int32)
		// Bytes 14-29: Part_name (16 bytes)

		partStart := int32(ebrBytes[2]) | int32(ebrBytes[3])<<8 | int32(ebrBytes[4])<<16 | int32(ebrBytes[5])<<24
		partSize := int32(ebrBytes[6]) | int32(ebrBytes[7])<<8 | int32(ebrBytes[8])<<16 | int32(ebrBytes[9])<<24
		nextEBR := int32(ebrBytes[10]) | int32(ebrBytes[11])<<8 | int32(ebrBytes[12])<<16 | int32(ebrBytes[13])<<24
		partName := strings.TrimRight(string(ebrBytes[14:30]), "\x00")

		if partSize == 0 {
			break
		}

		fit := "First Fit"
		if len(ebrBytes) > 2 {
			switch ebrBytes[1] {
			case 'B', 'b':
				fit = "Best Fit"
			case 'F', 'f':
				fit = "First Fit"
			case 'W', 'w':
				fit = "Worst Fit"
			}
		}

		logicalInfo := structs.PartitionInfo{
			Name:   partName,
			Type:   "Lógica",
			Status: "Activa",
			Fit:    fit,
			Start:  currentPos + partStart,
			Size:   partSize,
		}

		logicalPartitions = append(logicalPartitions, logicalInfo)

		if nextEBR == 0 {
			break
		}

		currentPos = extendedStart + nextEBR
	}

	return logicalPartitions, nil
}
