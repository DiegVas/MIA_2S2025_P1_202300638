package diskmanager

import (
	"fmt"
	"main/structs"
	"strings"
)

func FormatPartitionsOutput(diskInfo *structs.DiskPartitionsInfo) string {
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
