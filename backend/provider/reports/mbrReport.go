package reports

import (
	"encoding/binary"
	"fmt"

	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
)

// generateMBRReport genera el reporte del MBR y EBRs
func generateMBRReport(partitionID string) (string, error) {
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

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo con configuración más simple
	dot.WriteString("digraph MBR_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=box, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\"];\n")
	dot.WriteString("    node [fontname=\"Arial\"];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	diskName := getDiskNameFromPath(mountedPartition.Path)
	dot.WriteString(fmt.Sprintf("    title [label=\"MBR Report for %s\", fillcolor=lightblue];\n\n", diskName))

	// Tabla del MBR usando formato simple
	dot.WriteString(fmt.Sprintf("    mbr_table [label=\"MBR\\nSize: %d\\nPartitions: %d\\nFit: %d\", fillcolor=lightblue];\n\n",
		mbr.Mbr_size, mbr.Mbr_partitions, mbr.Dsk_fit[0]))

	// Tablas de las particiones usando formato record
	for i, partition := range mbr.Partitions {
		if partition.Part_status[0] != 0 { // Solo mostrar particiones activas
			partitionName := cleanString(string(partition.Part_name[:]))
			if partitionName == "" {
				partitionName = "unnamed"
			}

			// Usar formato de nodo simple
			dot.WriteString(fmt.Sprintf("    partition_%d [label=\"Partition %d\\nName: %s\\nStart: %d\\nSize: %d\", fillcolor=lightgreen];\n\n",
				i, i+1, partitionName, partition.Part_start, partition.Part_size))

			// Si es partición extendida, leer EBRs
			if partition.Part_type[0] == 'E' || partition.Part_type[0] == 'e' {
				ebrTables := generateEBRTables(mountedPartition.Path, partition.Part_start)
				dot.WriteString(ebrTables)
			}
		}
	}

	// Conectar elementos
	dot.WriteString("    title -> mbr_table;\n")
	for i, partition := range mbr.Partitions {
		if partition.Part_status[0] != 0 {
			dot.WriteString(fmt.Sprintf("    mbr_table -> partition_%d;\n", i))
		}
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

// cleanString limpia una cadena de caracteres no imprimibles
func cleanString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 { // Solo caracteres ASCII imprimibles
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// readMBRFromDisk lee el MBR desde el archivo del disco
func readMBRFromDisk(diskPath string) (*structs.MBR, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo del disco: %v", err)
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return nil, fmt.Errorf("error leyendo MBR: %v", err)
	}

	return &mbr, nil
}

// generateEBRTables genera las tablas de EBRs para particiones extendidas
func generateEBRTables(diskPath string, extendedStart int32) string {
	var dot strings.Builder

	file, err := os.Open(diskPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Leer EBRs en la partición extendida
	currentPos := extendedStart
	ebrCount := 0

	for currentPos != 0 && ebrCount < 10 { // Límite de seguridad
		// Posicionarse en el EBR
		file.Seek(int64(currentPos), 0)

		var ebr structs.EBR
		if err := binary.Read(file, binary.LittleEndian, &ebr); err != nil {
			break
		}

		// Solo mostrar EBRs con particiones válidas
		if ebr.Part_mount[0] != 0 {
			ebrName := cleanString(string(ebr.Part_name[:]))
			if ebrName == "" {
				ebrName = "unnamed"
			}

			// Convertir valores a representación segura
			mount := fmt.Sprintf("%d", ebr.Part_mount[0])
			fit := fmt.Sprintf("%d", ebr.Part_fit[0])

			// Usar formato simple
			dot.WriteString(fmt.Sprintf("    ebr_%d [label=\"EBR %d\\n", ebrCount, ebrCount+1))
			dot.WriteString(fmt.Sprintf("Mount: %s\\n", mount))
			dot.WriteString(fmt.Sprintf("Fit: %s\\n", fit))
			dot.WriteString(fmt.Sprintf("Start: %d\\n", ebr.Part_start))
			dot.WriteString(fmt.Sprintf("Size: %d\\n", ebr.Part_size))
			dot.WriteString(fmt.Sprintf("Next: %d\\n", ebr.Part_next))
			dot.WriteString(fmt.Sprintf("Name: %s\"", ebrName))
			dot.WriteString(", shape=box, style=filled, fillcolor=lightyellow];\n\n")
		} // Siguiente EBR
		if ebr.Part_next == 0 {
			break
		}
		currentPos = ebr.Part_next
		ebrCount++
	}

	return dot.String()
}

// getDiskNameFromPath extrae el nombre del disco de la ruta
func getDiskNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "disco.dsk"
}
