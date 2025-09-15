package reports

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
)

func generateTreeReport(partitionID string) (string, error) {
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

	// Abrir el archivo del disco
	file, err := os.Open(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Tree_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=plaintext];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	dot.WriteString(fmt.Sprintf("    title [label=\"Reporte Arbol del Sistema de Archivos\\nParticion: %s\", shape=box, style=filled, fillcolor=lightblue];\n\n", partitionID))

	// Generar el árbol completo del sistema de archivos
	err = generateFileSystemTree(file, superblock, &dot)
	if err != nil {
		return "", fmt.Errorf("error generando árbol del sistema de archivos: %v", err)
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

func generateFileSystemTree(file *os.File, superblock *structs.Superblock, dot *strings.Builder) error {
	// Leer bitmap de inodos para identificar cuáles están en uso
	inodeBitmap, err := readInodeBitmapFromFile(file, superblock)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	// Leer bitmap de bloques
	blockBitmap, err := readBlockBitmapFromFile(file, superblock)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	// Contador para nodos únicos
	nodeCounter := 0

	// Generar inodos utilizados
	err = generateUsedInodes(file, superblock, inodeBitmap, dot, &nodeCounter)
	if err != nil {
		return fmt.Errorf("error generando inodos: %v", err)
	}

	// Generar bloques utilizados
	err = generateUsedBlocks(file, superblock, blockBitmap, dot, &nodeCounter)
	if err != nil {
		return fmt.Errorf("error generando bloques: %v", err)
	}

	// Generar conexiones entre inodos y bloques
	err = generateInodeBlockConnections(file, superblock, inodeBitmap, blockBitmap, dot)
	if err != nil {
		return fmt.Errorf("error generando conexiones: %v", err)
	}

	// Generar conexiones desde bloques de directorio hacia inodos
	err = generateDirectoryBlockToInodeConnections(file, superblock, inodeBitmap, blockBitmap, dot)
	if err != nil {
		return fmt.Errorf("error generando conexiones de directorio a inodo: %v", err)
	}

	return nil
}
