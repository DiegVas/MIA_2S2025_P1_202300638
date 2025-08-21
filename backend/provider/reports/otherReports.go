package reports

import (
	"fmt"
	"main/provider/mounted"
	"strings"
)

// generateDiskReport genera el reporte del disco completo
func generateDiskReport(partitionID string) (string, error) {
	mountedPartition, err := mounted.GetMountedPartition(partitionID)
	if err != nil {
		return "", fmt.Errorf("error obteniendo partición montada: %v", err)
	}

	var dot strings.Builder
	dot.WriteString("digraph Disk_Report {\n")
	dot.WriteString("    rankdir=LR;\n")
	dot.WriteString("    node [shape=box];\n")

	diskName := getDiskNameFromPath(mountedPartition.DiskPath)
	dot.WriteString(fmt.Sprintf("    \"%s\" [label=\"Disco: %s\\nRuta: %s\"];\n",
		diskName, diskName, mountedPartition.DiskPath))

	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateInodeReport genera el reporte de inodos
func generateInodeReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Inode_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"Inode Report\" [label=\"Reporte de Inodos\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateBlockReport genera el reporte de bloques
func generateBlockReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Block_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"Block Report\" [label=\"Reporte de Bloques\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateBMInodeReport genera el reporte del bitmap de inodos
func generateBMInodeReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph BM_Inode_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"BM Inode Report\" [label=\"Bitmap de Inodos\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateBMBlockReport genera el reporte del bitmap de bloques
func generateBMBlockReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph BM_Block_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"BM Block Report\" [label=\"Bitmap de Bloques\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateTreeReport genera el reporte del árbol de directorios
func generateTreeReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Tree_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=box];\n")
	dot.WriteString("    root [label=\"/\"];\n")
	dot.WriteString("    \"Tree Report\" [label=\"Árbol de Directorios\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("    \"Tree Report\" -> root;\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateSuperblockReport genera el reporte del superbloque
func generateSuperblockReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Superblock_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"Superblock Report\" [label=\"Superbloque\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateFileReport genera el reporte de un archivo específico
func generateFileReport(partitionID, filePath string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph File_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString(fmt.Sprintf("    \"File Report\" [label=\"Archivo: %s\\nPartición: %s\"];\n",
		filePath, partitionID))
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateLSReport genera el reporte de listado de directorio
func generateLSReport(partitionID, dirPath string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph LS_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString(fmt.Sprintf("    \"LS Report\" [label=\"Listado: %s\\nPartición: %s\"];\n",
		dirPath, partitionID))
	dot.WriteString("}\n")
	return dot.String(), nil
}
