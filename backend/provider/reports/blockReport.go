package reports

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
)

func generateBlockReport(partitionID string) (string, error) {

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

	// Leer bitmap de bloques para identificar cuáles están en uso
	blockBitmap, err := readBlockBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	// Leer todos los bloques utilizados
	usedBlocks, err := readUsedBlocks(mountedPartition.Path, superblock, blockBitmap)
	if err != nil {
		return "", fmt.Errorf("error leyendo bloques: %v", err)
	}

	if len(usedBlocks) == 0 {
		return "", fmt.Errorf("no se encontraron bloques utilizados")
	}

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Block_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	titleLabel := fmt.Sprintf("BLOCK Report\\nParticion: %s", escapeLabel(partitionID))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Generar nodos para cada bloque utilizado
	for i, block := range usedBlocks {
		blockLabel := generateBlockLabel(i, block)
		dot.WriteString(fmt.Sprintf("    block_%d [label=\"%s\", fillcolor=lightyellow];\n", i, blockLabel))
	}

	// Generar conexiones para crear la lista enlazada
	dot.WriteString("\n    // Conexiones de la lista enlazada\n")
	dot.WriteString("    title -> block_0;\n")

	for i := 0; i < len(usedBlocks)-1; i++ {
		dot.WriteString(fmt.Sprintf("    block_%d -> block_%d;\n", i, i+1))
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

func readUsedBlocks(diskPath string, superblock *structs.Superblock, bitmap []byte) ([]blockInfo, error) {

	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var usedBlocks []blockInfo
	usedCount := 0

	for i := int32(0); i < superblock.S_blocks_count; i++ {
		if bitmap[i] == 1 { // Bloque en uso
			usedCount++

			// Calcular la posición correcta del bloque
			// Si S_block_start es 0, calcularlo manualmente
			var blockStartPos int32
			if superblock.S_block_start == 0 {
				// Calcular: inode_start + (inodes_count * inode_size)
				blockStartPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				if i == 0 {
				}
			} else {
				blockStartPos = superblock.S_block_start
			}

			blockPos := blockStartPos + (i * structs.BLOCK_SIZE)

			// Leer el bloque
			blockData := make([]byte, structs.BLOCK_SIZE)
			_, err := file.ReadAt(blockData, int64(blockPos))
			if err != nil {
				continue // Saltar bloques con error
			}

			// Determinar el tipo de bloque y leer su contenido
			blockType, content := determineBlockType(blockData)

			usedBlocks = append(usedBlocks, blockInfo{
				Number:    i,
				BlockType: blockType,
				Content:   content,
			})
		}
	}

	return usedBlocks, nil
}

// generateUsedBlocks genera los nodos de bloques utilizados
func generateUsedBlocks(file *os.File, superblock *structs.Superblock, bitmap []byte, dot *strings.Builder, nodeCounter *int) error {

	// Recorrer el bitmap para encontrar bloques utilizados
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		if i < int32(len(bitmap)) && bitmap[i] == 1 {

			// Este bloque está en uso, leerlo
			blockData, blockType, err := readBlockAt(file, i, superblock)
			if err != nil {
				fmt.Printf("DEBUG TREE: Error leyendo bloque %d: %v\n", i, err)
				continue // Saltar este bloque si hay error
			}

			// Generar tabla HTML para el bloque
			blockTable := generateBlockTableHTML(i, blockData, blockType)

			dot.WriteString(fmt.Sprintf("    block_%d [label=<%s>];\n", i, blockTable))
			(*nodeCounter)++
		}
	}

	return nil
}

func readBlockAt(file *os.File, blockNum int32, superblock *structs.Superblock) ([]byte, string, error) {
	blockData := make([]byte, structs.BLOCK_SIZE)

	var blockPos int32
	if superblock.S_block_start == 0 {
		blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
		blockPos += (blockNum * structs.BLOCK_SIZE)
	} else {
		blockPos = superblock.S_block_start + (blockNum * structs.BLOCK_SIZE)
	}

	_, err := file.ReadAt(blockData, int64(blockPos))
	if err != nil {
		return nil, "", err
	}

	// Determinar tipo de bloque basado en su contenido
	blockType := determineBlockTypeForTree(blockData)

	return blockData, blockType, nil
}
