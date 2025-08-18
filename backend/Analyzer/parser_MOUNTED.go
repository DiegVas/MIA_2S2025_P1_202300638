package Analyzer

import (
	"fmt"
	"main/structs"
	"strings"
)

func ParseMounted(command string) (bool, error) {

	command = strings.TrimSpace(strings.ToLower(command))

	if command != "mounted" {
		return false, fmt.Errorf("comando inválido: debe ser exactamente 'mounted'")
	}

	return true, nil
}

func GetMountedPartitions() []structs.MountedPartition {

	mountManager := structs.GetMountManager()

	return mountManager.GetMountedPartitions()
}

func FormatMountedPartitions() string {
	partitions := GetMountedPartitions()

	if len(partitions) == 0 {
		return "No hay particiones montadas actualmente en memoria"
	}

	var result strings.Builder
	result.WriteString("Particiones montadas en memoria:\n\n ")

	for i, partition := range partitions {
		result.WriteString(fmt.Sprintf("%d. ID: %s | Nombre: %s | Ruta: %s\n ",
			i+1, partition.ID, partition.Name, partition.Path))
	}

	return result.String()
}
