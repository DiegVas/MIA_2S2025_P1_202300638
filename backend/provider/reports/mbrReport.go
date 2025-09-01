package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

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

	// Formatear información del MBR
	diskSize := formatDiskSize(mbr.Mbr_size)
	creationDate := formatTimestamp(mbr.Mbr_timestamp)
	// diskSignature := formatDiskSignature(mbr.Mbr_uuid)  // Comentado temporalmente para evitar problemas DOT
	// fitType := formatFitType(mbr.Dsk_fit[0])            // Comentado temporalmente para evitar problemas DOT

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo con configuración más simple
	dot.WriteString("digraph MBR_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=box, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	diskName := getDiskNameFromPath(mountedPartition.Path)
	titleLabel := fmt.Sprintf("MBR Report\\n%s", escapeLabel(diskName))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Información detallada del MBR - simplificada para evitar problemas de escape
	mbrInfoLabel := fmt.Sprintf("MBR Information\\nDisk Size: %s\\nCreation Date: %s\\nTotal Partitions: %d",
		escapeLabel(diskSize), escapeLabel(creationDate), mbr.Mbr_partitions)

	dot.WriteString(fmt.Sprintf("    mbr_info [label=\"%s\", fillcolor=lightyellow];\n\n", mbrInfoLabel))

	// Analizar y organizar todas las particiones por tipo
	primaryPartitions := make([]structs.PartitionInfo, 0)
	extendedPartitions := make([]structs.PartitionInfo, 0)
	logicalPartitions := make([]structs.PartitionInfo, 0)

	// Procesar particiones primarias y extendidas del MBR
	for i, partition := range mbr.Partitions {
		// Condición: particiones activas (status == 1) O (status > 0 y tiene tipo válido)
		isActive := partition.Part_status[0] == 1 ||
			(partition.Part_status[0] > 0 && (partition.Part_type[0] == 'P' || partition.Part_type[0] == 'E' || partition.Part_type[0] == 'L'))

		if isActive && partition.Part_size > 0 {
			partInfo := structs.PartitionInfo{
				Name:   cleanString(string(partition.Part_name[:])),
				Type:   formatPartitionType(partition.Part_type[0]),
				Status: formatPartitionStatus(partition.Part_status[0]),
				Fit:    formatFitType(partition.Part_fit[0]),
				Start:  partition.Part_start,
				Size:   partition.Part_size,
			}

			if partInfo.Name == "" {
				partInfo.Name = fmt.Sprintf("partition_%d", i+1)
			}

			switch partition.Part_type[0] {
			case 'P':
				primaryPartitions = append(primaryPartitions, partInfo)
			case 'E':
				extendedPartitions = append(extendedPartitions, partInfo)
				// Leer particiones lógicas de la partición extendida
				logicals := readLogicalPartitionsFromEBR(mountedPartition.Path, partition.Part_start)
				logicalPartitions = append(logicalPartitions, logicals...)
			case 'L':
				// Esto no debería ocurrir en el MBR, las lógicas están en EBRs
				partInfo.Type = "Logical (error: en MBR)"
				primaryPartitions = append(primaryPartitions, partInfo)
			}
		}
	}

	// Crear secciones para cada tipo de partición
	dotContent := generatePartitionSections(primaryPartitions, extendedPartitions, logicalPartitions)
	dot.WriteString(dotContent)

	// Conexiones principales
	dot.WriteString("    title -> mbr_info;\n")

	if len(primaryPartitions) > 0 {
		dot.WriteString("    mbr_info -> primary_section;\n")
	}
	if len(extendedPartitions) > 0 {
		dot.WriteString("    mbr_info -> extended_section;\n")
	}
	if len(logicalPartitions) > 0 {
		dot.WriteString("    mbr_info -> logical_section;\n")
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

// readMBRFromDisk lee el MBR desde el archivo del disco usando el mismo método que disk_partition.go
func readMBRFromDisk(diskPath string) (*structs.MBR, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo del disco: %v", err)
	}
	defer file.Close()

	// Usar la misma lógica que disk_partition.go
	const MBR_SIZE = 47

	mbrBytes := make([]byte, MBR_SIZE)
	_, err = file.ReadAt(mbrBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR: %v", err)
	}

	var mbr structs.MBR
	buffer := bytes.NewReader(mbrBytes)

	// Leer cabecera del MBR
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_size)
	buffer.Read(mbr.Mbr_timestamp[:])
	buffer.Read(mbr.Mbr_uuid[:])
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_partitions)
	buffer.Read(mbr.Dsk_fit[:])

	// Leer particiones por separado
	partitionDataOffset := int64(MBR_SIZE)
	partitionSize := int64(59) // Tamaño de cada estructura PARTITION

	for i := 0; i < 4; i++ {
		partPos := partitionDataOffset + int64(i)*partitionSize

		partData := make([]byte, partitionSize)
		_, err = file.ReadAt(partData, partPos)
		if err != nil {
			return nil, fmt.Errorf("error al leer la partición %d: %v", i, err)
		}

		partBuffer := bytes.NewReader(partData)

		partBuffer.Read(mbr.Partitions[i].Part_status[:])
		partBuffer.Read(mbr.Partitions[i].Part_type[:])
		partBuffer.Read(mbr.Partitions[i].Part_fit[:])
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_start)
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_size)
		partBuffer.Read(mbr.Partitions[i].Part_name[:])
		partBuffer.Read(mbr.Partitions[i].Part_uuid[:])
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_correlative)
	}

	return &mbr, nil
}

// getDiskNameFromPath extrae el nombre del disco de la ruta
func getDiskNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "disco.dsk"
}

// formatDiskSize convierte el tamaño en bytes a una representación legible
func formatDiskSize(sizeBytes int32) string {
	size := float64(sizeBytes)
	units := []string{"B", "KB", "MB", "GB"}

	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f %s", size, unit)
			}
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%.2f GB", size)
}

// formatTimestamp convierte el timestamp del MBR a una fecha legible
func formatTimestamp(timestamp [10]byte) string {
	// El timestamp está en formato YYYY-MM-DD como string
	timestampStr := string(timestamp[:])
	// Limpiar caracteres nulos
	timestampStr = strings.TrimRight(timestampStr, "\x00")

	if len(timestampStr) >= 10 {
		// Intentar parsear la fecha
		if t, err := time.Parse("2006-01-02", timestampStr); err == nil {
			return t.Format("January 2, 2006")
		}
	}

	// Si no se puede parsear, devolver tal como está
	if timestampStr == "" {
		return "Unknown"
	}
	return timestampStr
}

// formatFitType convierte el código de fit a una descripción legible
func formatFitType(fitByte byte) string {
	switch fitByte {
	case 'F', 'f':
		return "First Fit"
	case 'B', 'b':
		return "Best Fit"
	case 'W', 'w':
		return "Worst Fit"
	default:
		return fmt.Sprintf("Unknown (%c)", fitByte)
	}
}

// formatPartitionType convierte el código de tipo de partición a descripción legible
func formatPartitionType(typeByte byte) string {
	switch typeByte {
	case 'P', 'p':
		return "Primary"
	case 'E', 'e':
		return "Extended"
	case 'L', 'l':
		return "Logical"
	default:
		return fmt.Sprintf("Unknown (%c)", typeByte)
	}
}

// formatPartitionStatus convierte el código de estado de partición a descripción legible
func formatPartitionStatus(statusByte byte) string {
	switch statusByte {
	case 'A', 'a':
		return "Active"
	case 1:
		return "Active"
	case 0:
		return "Inactive"
	default:
		return fmt.Sprintf("Unknown (%d)", statusByte)
	}
}

// readLogicalPartitionsFromEBR lee las particiones lógicas de una partición extendida
func readLogicalPartitionsFromEBR(diskPath string, extendedStart int32) []structs.PartitionInfo {
	logicalPartitions := make([]structs.PartitionInfo, 0)

	file, err := os.Open(diskPath)
	if err != nil {
		return logicalPartitions
	}
	defer file.Close()

	currentPos := extendedStart
	logicalIndex := 1
	ebrSize := int64(30) // Tamaño del EBR según disk_partition.go

	for currentPos != 0 && logicalIndex <= 50 { // Límite de seguridad
		// Leer datos del EBR usando la misma lógica que disk_partition.go
		ebrData := make([]byte, ebrSize)
		_, err := file.ReadAt(ebrData, int64(currentPos))
		if err != nil {
			break
		}

		// Extraer datos del EBR manualmente (como en disk_partition.go)
		// partMount := ebrData[0]  // No necesario para el reporte
		partFit := ebrData[1]
		partStart := int32(ebrData[2]) | int32(ebrData[3])<<8 | int32(ebrData[4])<<16 | int32(ebrData[5])<<24
		partSize := int32(ebrData[6]) | int32(ebrData[7])<<8 | int32(ebrData[8])<<16 | int32(ebrData[9])<<24
		partNext := int32(ebrData[10]) | int32(ebrData[11])<<8 | int32(ebrData[12])<<16 | int32(ebrData[13])<<24
		partName := strings.TrimRight(string(ebrData[14:30]), "\x00")

		// Solo procesar EBRs con particiones válidas (tamaño > 0)
		if partSize > 0 {
			partInfo := structs.PartitionInfo{
				Name:   cleanString(partName),
				Type:   "Logical",
				Status: "Active",
				Fit:    formatFitType(partFit),
				Start:  currentPos + partStart, // Start relativo a la posición del EBR
				Size:   partSize,
			}

			if partInfo.Name == "" {
				partInfo.Name = fmt.Sprintf("logical_%d", logicalIndex)
			}

			logicalPartitions = append(logicalPartitions, partInfo)
			logicalIndex++
		}

		// Buscar el siguiente EBR
		if partNext == 0 {
			// No hay más EBRs
			break
		}

		// Part_next se guarda como relativo al inicio de la partición extendida
		nextPos := extendedStart + partNext

		// Verificar que la posición sea válida
		if nextPos <= currentPos || nextPos >= extendedStart+10485760 {
			break
		}

		currentPos = nextPos
	}

	return logicalPartitions
} // generatePartitionSections genera las secciones organizadas por tipo de partición
func generatePartitionSections(primary, extended, logical []structs.PartitionInfo) string {
	var dot strings.Builder

	// Sección de particiones primarias
	if len(primary) > 0 {
		dot.WriteString("    primary_section [label=\"PRIMARY PARTITIONS\", shape=box, style=filled, fillcolor=lightcyan];\n")
		for i, part := range primary {
			nodeId := fmt.Sprintf("primary_%d", i)
			partSize := formatDiskSize(part.Size)
			// Escapar caracteres especiales en las etiquetas
			partName := escapeLabel(part.Name)
			partType := escapeLabel(part.Type)
			partStatus := escapeLabel(part.Status)
			partFit := escapeLabel(part.Fit)

			dot.WriteString(fmt.Sprintf("    %s [label=\"%s\\nType: %s\\nStatus: %s\\nStart: %d\\nSize: %s\\nFit: %s\", fillcolor=lightgreen];\n",
				nodeId, partName, partType, partStatus, part.Start, partSize, partFit))
			dot.WriteString(fmt.Sprintf("    primary_section -> %s;\n", nodeId))
		}
		dot.WriteString("\n")
	}

	// Sección de particiones extendidas
	if len(extended) > 0 {
		dot.WriteString("    extended_section [label=\"EXTENDED PARTITIONS\", shape=box, style=filled, fillcolor=lightyellow];\n")
		for i, part := range extended {
			nodeId := fmt.Sprintf("extended_%d", i)
			partSize := formatDiskSize(part.Size)
			// Escapar caracteres especiales en las etiquetas
			partName := escapeLabel(part.Name)
			partType := escapeLabel(part.Type)
			partStatus := escapeLabel(part.Status)
			partFit := escapeLabel(part.Fit)

			dot.WriteString(fmt.Sprintf("    %s [label=\"%s\\nType: %s\\nStatus: %s\\nStart: %d\\nSize: %s\\nFit: %s\", fillcolor=orange];\n",
				nodeId, partName, partType, partStatus, part.Start, partSize, partFit))
			dot.WriteString(fmt.Sprintf("    extended_section -> %s;\n", nodeId))
		}
		dot.WriteString("\n")
	}

	// Sección de particiones lógicas
	if len(logical) > 0 {
		dot.WriteString("    logical_section [label=\"LOGICAL PARTITIONS\", shape=box, style=filled, fillcolor=lightpink];\n")
		for i, part := range logical {
			nodeId := fmt.Sprintf("logical_%d", i)
			partSize := formatDiskSize(part.Size)
			// Escapar caracteres especiales en las etiquetas
			partName := escapeLabel(part.Name)
			partType := escapeLabel(part.Type)
			partStatus := escapeLabel(part.Status)
			partFit := escapeLabel(part.Fit)

			dot.WriteString(fmt.Sprintf("    %s [label=\"%s\\nType: %s\\nStatus: %s\\nStart: %d\\nSize: %s\\nFit: %s\", fillcolor=lightcoral];\n",
				nodeId, partName, partType, partStatus, part.Start, partSize, partFit))
			dot.WriteString(fmt.Sprintf("    logical_section -> %s;\n", nodeId))
		}
		dot.WriteString("\n")
	}

	return dot.String()
}

// escapeLabel escapa caracteres especiales para etiquetas DOT
func escapeLabel(s string) string {
	// Reemplazar caracteres problemáticos para DOT
	s = strings.ReplaceAll(s, "\\", "\\\\") // Escapar backslashes primero
	s = strings.ReplaceAll(s, "\"", "\\\"") // Escapar comillas dobles
	s = strings.ReplaceAll(s, "\n", "\\n")  // Escapar saltos de línea
	s = strings.ReplaceAll(s, "\t", "\\t")  // Escapar tabs
	s = strings.ReplaceAll(s, "\r", "\\r")  // Escapar retorno de carro

	// Remover caracteres de control y no imprimibles
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 { // Solo caracteres ASCII imprimibles
			result.WriteRune(r)
		} else if r == '\n' || r == '\t' {
			result.WriteRune(r) // Permitir \n y \t ya escapados
		}
	}
	s = result.String()

	// Limitar longitud para evitar problemas
	if len(s) > 100 {
		s = s[:97] + "..."
	}

	// Si la cadena está vacía después del filtrado, usar un placeholder
	if strings.TrimSpace(s) == "" {
		return "N/A"
	}

	return s
}
