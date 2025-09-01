package reports

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

// DiskElement representa un elemento del disco (MBR, partición, espacio libre)
type DiskElement struct {
	Name        string
	Type        string // "MBR", "Primaria", "Extendida", "Lógica", "Libre"
	Start       int32
	Size        int32
	Percentage  float64
	SubElements []DiskElement // Para particiones lógicas dentro de extendida
}

// generateDiskReport genera el reporte del disco con todos los porcentajes
func generateDiskReport(partitionID string) (string, error) {
	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el MBR del disco
	mbr, err := utils.ReadMBR(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Analizar la estructura del disco
	diskElements, err := analyzeDiskStructure(mountedPartition.Path, mbr)
	if err != nil {
		return "", fmt.Errorf("error analizando estructura del disco: %v", err)
	}

	// Generar contenido DOT con tabla completa
	dotContent := generateDiskTable(diskElements, mbr.Mbr_size)

	return dotContent, nil
}

// analyzeDiskStructure analiza toda la estructura del disco y calcula porcentajes
func analyzeDiskStructure(diskPath string, mbr *structs.MBR) ([]DiskElement, error) {
	var elements []DiskElement
	totalSize := mbr.Mbr_size
	currentPos := int32(283) // Después del MBR y las 4 entradas de partición (47 + 4*59)

	// Elemento del MBR
	mbrElement := DiskElement{
		Name:       "MBR",
		Type:       "MBR",
		Start:      0,
		Size:       283,
		Percentage: float64(283) / float64(totalSize) * 100,
	}
	elements = append(elements, mbrElement)

	// Procesar las 4 particiones en orden
	for i := 0; i < 4; i++ {
		partition := &mbr.Partitions[i]

		// Saltar particiones vacías
		if partition.Part_size == 0 {
			continue
		}

		// Agregar espacio libre antes de la partición si existe
		if partition.Part_start > currentPos {
			freeSize := partition.Part_start - currentPos
			freeElement := DiskElement{
				Name:       "Libre",
				Type:       "Libre",
				Start:      currentPos,
				Size:       freeSize,
				Percentage: float64(freeSize) / float64(totalSize) * 100,
			}
			elements = append(elements, freeElement)
		}

		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		if partName == "" {
			partName = fmt.Sprintf("Partición_%d", i+1)
		}

		// Determinar tipo de partición
		if partition.Part_type[0] == 'E' || partition.Part_type[0] == 'e' {
			// Partición extendida
			extElement := DiskElement{
				Name:       partName,
				Type:       "Extendida",
				Start:      partition.Part_start,
				Size:       partition.Part_size,
				Percentage: float64(partition.Part_size) / float64(totalSize) * 100,
			}

			// Analizar particiones lógicas dentro de la extendida
			logicals, err := analyzeLogicalPartitions(diskPath, partition.Part_start, partition.Part_size, totalSize)
			if err == nil {
				extElement.SubElements = logicals
			}

			elements = append(elements, extElement)
		} else {
			// Partición primaria
			primElement := DiskElement{
				Name:       partName,
				Type:       "Primaria",
				Start:      partition.Part_start,
				Size:       partition.Part_size,
				Percentage: float64(partition.Part_size) / float64(totalSize) * 100,
			}
			elements = append(elements, primElement)
		}

		currentPos = partition.Part_start + partition.Part_size
	}

	// Agregar espacio libre al final si existe
	if currentPos < totalSize {
		freeSize := totalSize - currentPos
		freeElement := DiskElement{
			Name:       "Libre",
			Type:       "Libre",
			Start:      currentPos,
			Size:       freeSize,
			Percentage: float64(freeSize) / float64(totalSize) * 100,
		}
		elements = append(elements, freeElement)
	}

	return elements, nil
}

// analyzeLogicalPartitions analiza las particiones lógicas dentro de una extendida
func analyzeLogicalPartitions(diskPath string, extStart, extSize int32, totalDiskSize int32) ([]DiskElement, error) {
	var logicals []DiskElement

	file, err := os.Open(diskPath)
	if err != nil {
		return logicals, err
	}
	defer file.Close()

	currentPos := extStart
	ebrSize := int32(30) // Tamaño del EBR

	for {
		// Leer EBR
		ebrBytes := make([]byte, 30)
		_, err := file.ReadAt(ebrBytes, int64(currentPos))
		if err != nil {
			break
		}

		// Parsear EBR
		partStart := int32(ebrBytes[2]) | int32(ebrBytes[3])<<8 | int32(ebrBytes[4])<<16 | int32(ebrBytes[5])<<24
		partSize := int32(ebrBytes[6]) | int32(ebrBytes[7])<<8 | int32(ebrBytes[8])<<16 | int32(ebrBytes[9])<<24
		nextEBR := int32(ebrBytes[10]) | int32(ebrBytes[11])<<8 | int32(ebrBytes[12])<<16 | int32(ebrBytes[13])<<24
		partName := strings.TrimRight(string(ebrBytes[14:30]), "\x00")

		if partSize == 0 {
			break
		}

		if partName == "" {
			partName = "Lógica"
		}

		// Agregar EBR
		ebrElement := DiskElement{
			Name:       "EBR",
			Type:       "EBR",
			Start:      currentPos,
			Size:       ebrSize,
			Percentage: float64(ebrSize) / float64(totalDiskSize) * 100,
		}
		logicals = append(logicals, ebrElement)

		// Agregar partición lógica
		logElement := DiskElement{
			Name:       partName,
			Type:       "Lógica",
			Start:      currentPos + partStart,
			Size:       partSize,
			Percentage: float64(partSize) / float64(totalDiskSize) * 100,
		}
		logicals = append(logicals, logElement)

		// Calcular espacio libre dentro de esta sección
		nextSectionStart := currentPos + partStart + partSize
		if nextEBR != 0 {
			nextEBRPos := extStart + nextEBR
			if nextEBRPos > nextSectionStart {
				freeSize := nextEBRPos - nextSectionStart
				freeElement := DiskElement{
					Name:       "Libre",
					Type:       "Libre",
					Start:      nextSectionStart,
					Size:       freeSize,
					Percentage: float64(freeSize) / float64(totalDiskSize) * 100,
				}
				logicals = append(logicals, freeElement)
			}
			currentPos = nextEBRPos
		} else {
			// Último EBR, verificar espacio libre hasta el final de la extendida
			extEnd := extStart + extSize
			if nextSectionStart < extEnd {
				freeSize := extEnd - nextSectionStart
				freeElement := DiskElement{
					Name:       "Libre",
					Type:       "Libre",
					Start:      nextSectionStart,
					Size:       freeSize,
					Percentage: float64(freeSize) / float64(totalDiskSize) * 100,
				}
				logicals = append(logicals, freeElement)
			}
			break
		}
	}

	return logicals, nil
}

// generateDiskTable genera la tabla HTML con la estructura del disco usando la plantilla base
func generateDiskTable(elements []DiskElement, totalSize int32) string {
	// Separar elementos por tipo
	var mbrContent, logicaContent, extendidaContent strings.Builder

	for _, element := range elements {
		elementInfo := fmt.Sprintf("%s: %.2f%%<br/>", element.Name, element.Percentage)

		switch element.Type {
		case "MBR":
			mbrContent.WriteString(elementInfo)
		case "Primaria":
			mbrContent.WriteString(elementInfo)
		case "Extendida":
			extendidaContent.WriteString(elementInfo)
			// TODAS las particiones lógicas y EBRs van en la columna Extendida
			for _, subElement := range element.SubElements {
				subInfo := fmt.Sprintf("└─ %s: %.2f%%<br/>", subElement.Name, subElement.Percentage)
				extendidaContent.WriteString(subInfo)
			}
		case "Libre":
			mbrContent.WriteString(elementInfo)
		}
	}

	// Usar la plantilla base corregida
	dotContent := fmt.Sprintf(`digraph G {
    node [shape=plaintext]

    tabla [
        label=<
            <table BORDER="1" CELLBORDER="1" CELLSPACING="0">
                <tr>
                    <td bgcolor="lightgrey"><b>MBR</b></td>
                    <td bgcolor="lightgrey"><b>Lógica</b></td>
                    <td bgcolor="lightgrey"><b>Extendida</b></td>
                </tr>
                <tr>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                </tr>
            </table>
        >
    ];
}`, mbrContent.String(), logicaContent.String(), extendidaContent.String())

	return dotContent
}

// getColorForType retorna el color correspondiente a cada tipo de elemento
func getColorForType(elementType string) string {
	switch elementType {
	case "MBR":
		return "lightcoral"
	case "Primaria":
		return "lightgreen"
	case "Extendida":
		return "lightyellow"
	case "Lógica":
		return "lightpink"
	case "EBR":
		return "orange"
	case "Libre":
		return "lightgray"
	default:
		return "white"
	}
}
