package Analyzer

import (
	"fmt"
	"main/utils"
	"os"
	"strings"
)

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

func ShowDiskPartitions(params *PdiskParams) (*DiskPartitionsInfo, error) {
	if _, err := os.Stat(params.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo del disco '%s' no existe", params.Path)
	}

	mbr, err := utils.ReadMBR(params.Path)
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	diskInfo := &DiskPartitionsInfo{
		DiskPath:   params.Path,
		Partitions: make([]PartitionInfo, 0),
	}

	for _, partition := range mbr.Partitions {
		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		if partName == "" {
			continue
		}

		status := "Inactiva"
		if len(partition.Part_status) > 0 && partition.Part_status[0] == 1 {
			status = "Activa"
		}

		partType := "Desconocido"
		if len(partition.Part_type) > 0 {
			switch partition.Part_type[0] {
			case 'P', 'p':
				partType = "Primaria"
			case 'E', 'e':
				partType = "Extendida"
			case 'L', 'l':
				partType = "Lógica"
			default:
				partType = fmt.Sprintf("Tipo_%c", partition.Part_type[0])
			}
		}

		fit := "Desconocido"
		if len(partition.Part_fit) > 0 {
			switch partition.Part_fit[0] {
			case 'B', 'b':
				fit = "Best Fit"
			case 'F', 'f':
				fit = "First Fit"
			case 'W', 'w':
				fit = "Worst Fit"
			default:
				fit = fmt.Sprintf("Fit_%c", partition.Part_fit[0])
			}
		}

		partInfo := PartitionInfo{
			Name:   partName,
			Type:   partType,
			Status: status,
			Fit:    fit,
			Start:  partition.Part_start,
			Size:   partition.Part_size,
		}

		diskInfo.Partitions = append(diskInfo.Partitions, partInfo)

		if partition.Part_type[0] == 'E' || partition.Part_type[0] == 'e' {
			logicalPartitions, err := readLogicalPartitions(params.Path, partition.Part_start)
			if err == nil {
				diskInfo.Partitions = append(diskInfo.Partitions, logicalPartitions...)
			}
		}
	}

	return diskInfo, nil
}

func readLogicalPartitions(diskPath string, extendedStart int32) ([]PartitionInfo, error) {
	var logicalPartitions []PartitionInfo

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

		logicalInfo := PartitionInfo{
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

func FormatPartitionsOutput(diskInfo *DiskPartitionsInfo) string {
	if len(diskInfo.Partitions) == 0 {
		return fmt.Sprintf("El disco '%s' no contiene particiones.", diskInfo.DiskPath)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("=== PARTICIONES DEL DISCO: %s ===\n\n", diskInfo.DiskPath))

	for i, partition := range diskInfo.Partitions {
		output.WriteString(fmt.Sprintf("%d. Partición %s:\n", i+1, partition.Name))
		output.WriteString(fmt.Sprintf("  Tipo: %s\n", partition.Type))
		output.WriteString(fmt.Sprintf("  Tamaño: %d bytes\n", partition.Size))

		output.WriteString("\n")
	}

	return output.String()
}
