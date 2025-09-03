package reports

import (
	"fmt"
	"main/provider/mounted"
	"strings"
)

// generateDiskReport genera el reporte DISK mostrando la estructura de particiones y porcentajes
func generateDiskReport(partitionID string) (string, error) {
	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el MBR del disco
	mbr, err := readMBRFromDisk(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Analizar todas las particiones y calcular porcentajes
	diskSize := float64(mbr.Mbr_size)
	var primaryPartitions []partitionInfo
	var extendedPartitions []extendedPartitionInfo
	var usedSpace float64

	// Procesar particiones del MBR
	for i, partition := range mbr.Partitions {
		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")

		// Si la partición tiene nombre y tamaño, está en uso
		if partName != "" && partition.Part_size > 0 {
			partType := string(partition.Part_type[:1])
			percentage := (float64(partition.Part_size) / diskSize) * 100

			switch partType {
			case "P", "p":
				// Partición primaria
				partInfo := partitionInfo{
					Name:       partName,
					Size:       partition.Part_size,
					Start:      partition.Part_start,
					Type:       "Primaria",
					Percentage: percentage,
					Index:      i,
				}
				primaryPartitions = append(primaryPartitions, partInfo)
				usedSpace += float64(partition.Part_size)

			case "E", "e":
				// Partición extendida
				logicalPartitions := readLogicalPartitionsFromEBR(mountedPartition.Path, partition.Part_start)

				// Calcular espacio usado por particiones lógicas
				var logicalUsedSpace float64
				var logicals []partitionInfo

				for _, logical := range logicalPartitions {
					logicalPercentage := (float64(logical.Size) / diskSize) * 100
					logicalInfo := partitionInfo{
						Name:       logical.Name,
						Size:       logical.Size,
						Start:      logical.Start,
						Type:       "Lógica",
						Percentage: logicalPercentage,
						Index:      -1, // Las lógicas no tienen índice en el MBR
					}
					logicals = append(logicals, logicalInfo)
					logicalUsedSpace += float64(logical.Size)
				}

				// Calcular espacio libre dentro de la partición extendida
				extendedFreeSpace := float64(partition.Part_size) - logicalUsedSpace
				extendedFreePercentage := (extendedFreeSpace / diskSize) * 100

				extInfo := extendedPartitionInfo{
					Name:                partName,
					Size:                partition.Part_size,
					Start:               partition.Part_start,
					Percentage:          percentage,
					LogicalPartitions:   logicals,
					FreeSpace:           int32(extendedFreeSpace),
					FreeSpacePercentage: extendedFreePercentage,
					Index:               i,
				}
				extendedPartitions = append(extendedPartitions, extInfo)
				usedSpace += float64(partition.Part_size)
			}
		}
	}

	// Calcular espacio libre del disco
	freeSpace := diskSize - usedSpace
	freePercentage := (freeSpace / diskSize) * 100

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph DISK_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n\n")

	// Título del reporte
	diskName := getDiskNameFromPath(mountedPartition.Path)
	titleLabel := fmt.Sprintf("DISK Report\\n%s", escapeLabel(diskName))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Información general del disco
	diskInfo := fmt.Sprintf("Tamaño Total: %s|Espacio Libre: %.2f%%",
		formatDiskSize(mbr.Mbr_size), freePercentage)
	dot.WriteString(fmt.Sprintf("    disk_info [label=\"%s\", fillcolor=lightgray];\n\n", diskInfo))

	// Tabla de estructura del disco
	dot.WriteString("    disk_structure [label=\"")

	// Header de la tabla
	dot.WriteString("Estructura del Disco|")
	dot.WriteString("{Componente|Tamaño|Porcentaje|Detalles}|")

	// MBR (no cuenta para el porcentaje pero se muestra)
	dot.WriteString("{MBR|47 bytes|N/A|Master Boot Record}|")

	// Particiones primarias
	for _, part := range primaryPartitions {
		dot.WriteString(fmt.Sprintf("{%s|%s|%.2f%%|Partición Primaria}|",
			escapeLabel(part.Name),
			formatDiskSize(part.Size),
			part.Percentage))
	}

	// Particiones extendidas y sus lógicas
	for _, ext := range extendedPartitions {
		dot.WriteString(fmt.Sprintf("{%s|%s|%.2f%%|Partición Extendida}|",
			escapeLabel(ext.Name),
			formatDiskSize(ext.Size),
			ext.Percentage))

		// Mostrar particiones lógicas dentro de la extendida
		for _, logical := range ext.LogicalPartitions {
			dot.WriteString(fmt.Sprintf("{  └─ %s|%s|%.2f%%|Partición Lógica}|",
				escapeLabel(logical.Name),
				formatDiskSize(logical.Size),
				logical.Percentage))
		}

		// Mostrar espacio libre de la extendida si existe
		if ext.FreeSpace > 0 {
			dot.WriteString(fmt.Sprintf("{  └─ Libre (Ext)|%s|%.2f%%|Espacio libre en extendida}|",
				formatDiskSize(ext.FreeSpace),
				ext.FreeSpacePercentage))
		}
	}

	// Espacio libre del disco
	if freeSpace > 0 {
		dot.WriteString(fmt.Sprintf("{Espacio Libre|%s|%.2f%%|Espacio no particionado}",
			formatDiskSize(int32(freeSpace)),
			freePercentage))
	}

	dot.WriteString("\", fillcolor=lightyellow];\n\n")

	// Conexiones
	dot.WriteString("    title -> disk_info;\n")
	dot.WriteString("    disk_info -> disk_structure;\n")

	dot.WriteString("}\n")

	return dot.String(), nil
}

// partitionInfo estructura auxiliar para manejar información de particiones
type partitionInfo struct {
	Name       string
	Size       int32
	Start      int32
	Type       string
	Percentage float64
	Index      int
}

// extendedPartitionInfo estructura auxiliar para manejar particiones extendidas
type extendedPartitionInfo struct {
	Name                string
	Size                int32
	Start               int32
	Percentage          float64
	LogicalPartitions   []partitionInfo
	FreeSpace           int32
	FreeSpacePercentage float64
	Index               int
}
