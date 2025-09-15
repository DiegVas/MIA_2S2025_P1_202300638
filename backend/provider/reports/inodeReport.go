package reports

import (
	"fmt"
	"main/provider/mounted"
	"strings"
)

// generateInodeReport genera el reporte de inodos utilizados
func generateInodeReport(partitionID string) (string, error) {

	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el superbloque para obtener información del sistema de archivos
	superblock, err := readSuperblock(mountedPartition.Path, partitionID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Leer bitmap de inodos para identificar cuáles están en uso
	inodeBitmap, err := readInodeBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	// Leer todos los inodos utilizados
	usedInodes, err := readUsedInodes(mountedPartition.Path, superblock, inodeBitmap)
	if err != nil {
		return "", fmt.Errorf("error leyendo inodos: %v", err)
	}

	if len(usedInodes) == 0 {
		return "", fmt.Errorf("no se encontraron inodos utilizados")
	}

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Inode_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	titleLabel := fmt.Sprintf("INODE Report\\nParticion: %s", escapeLabel(partitionID))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Generar nodos para cada inodo utilizado
	for i, inodeInfo := range usedInodes {
		inodeLabel := generateInodeLabel(i, inodeInfo)

		// Determinar color según el tipo de inodo
		var fillColor string
		if inodeInfo.Inode.I_type == 1 {
			fillColor = "lightgreen" // Directorios en verde claro
		} else {
			fillColor = "lightcyan" // Archivos en cian claro
		}

		dot.WriteString(fmt.Sprintf("    inode_%d [label=\"%s\", fillcolor=%s];\n", i, inodeLabel, fillColor))
	}

	// Generar conexiones para crear la lista enlazada
	dot.WriteString("\n    // Conexiones de la lista enlazada\n")
	dot.WriteString("    title -> inode_0;\n")

	for i := 0; i < len(usedInodes)-1; i++ {
		dot.WriteString(fmt.Sprintf("    inode_%d -> inode_%d;\n", i, i+1))
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}
