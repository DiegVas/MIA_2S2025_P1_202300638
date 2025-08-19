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

		partName := strings.TrimRight(string(ebrBytes[26:42]), "\x00")
		if partName == "" {
			break
		}

		var partStart, partSize, nextEBR int32

		partStart = int32(ebrBytes[8]) | int32(ebrBytes[9])<<8 | int32(ebrBytes[10])<<16 | int32(ebrBytes[11])<<24
		partSize = int32(ebrBytes[12]) | int32(ebrBytes[13])<<8 | int32(ebrBytes[14])<<16 | int32(ebrBytes[15])<<24
		nextEBR = int32(ebrBytes[16]) | int32(ebrBytes[17])<<8 | int32(ebrBytes[18])<<16 | int32(ebrBytes[19])<<24

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
