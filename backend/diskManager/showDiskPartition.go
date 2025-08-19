package diskmanager

import (
	"fmt"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

func ShowDiskPartitions(params *Analyzer.PdiskParams) (*structs.DiskPartitionsInfo, error) {
	if _, err := os.Stat(params.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo del disco '%s' no existe", params.Path)
	}

	mbr, err := utils.ReadMBR(params.Path)
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	diskInfo := &structs.DiskPartitionsInfo{
		DiskPath:   params.Path,
		Partitions: make([]structs.PartitionInfo, 0),
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

		partInfo := structs.PartitionInfo{
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
