package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func generateIndirectBlockConnections(file *os.File, superblock *structs.Superblock, blockNum int32, blockBitmap []byte, dot *strings.Builder) error {
	// Leer el bloque de punteros
	pointerBlock, err := readPointerBlockAt(file, blockNum, superblock)
	if err != nil {
		return err
	}

	// Conectar con cada bloque apuntado
	for i, pointer := range pointerBlock.B_pointers {
		// Validar que el puntero esté en un rango razonable y esté marcado como usado
		if pointer != 0 && pointer > 0 && pointer < 10000 && pointer < int32(len(blockBitmap)) && blockBitmap[pointer] == 1 {
			fmt.Printf("DEBUG TREE: Generando conexion indirecta: block_%d:ptr_%d -> block_%d\n", blockNum, i, pointer)
			dot.WriteString(fmt.Sprintf("    block_%d:ptr_%d -> block_%d;\n", blockNum, i, pointer))
		} else if pointer != 0 {
			fmt.Printf("DEBUG TREE: Puntero invalido omitido: %d (rango: 0-%d, bitmap len: %d)\n", pointer, 10000, len(blockBitmap))
		}
	}

	return nil
}

func generatePointerBlockTable(table *strings.Builder, blockData []byte) error {
	var pointerBlock structs.PointerBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err != nil {
		return err
	}

	ptrCount := 0
	for i, pointer := range pointerBlock.B_pointers {

		// Validar que el puntero esté en un rango razonable (por ejemplo, < 10000)
		if pointer != 0 && pointer < 10000 && ptrCount < 5 { // Mostrar solo los primeros 5 punteros válidos
			table.WriteString(fmt.Sprintf("<tr><td port=\"ptr_%d\">Puntero %d</td><td>%d</td></tr>", i, i, pointer))
			ptrCount++
		}
	}

	// Contar total de punteros válidos
	totalValid := 0
	for _, pointer := range pointerBlock.B_pointers {
		if pointer != 0 && pointer < 10000 { // Solo contar punteros válidos
			totalValid++
		}
	}

	if totalValid > 5 {
		table.WriteString(fmt.Sprintf("<tr><td colspan=\"2\">... y %d punteros mas</td></tr>", totalValid-5))
	}

	return nil
}

// readPointerBlockAt lee un bloque de punteros específico
func readPointerBlockAt(file *os.File, blockNum int32, superblock *structs.Superblock) (*structs.PointerBlock, error) {
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
		return nil, err
	}

	var pointerBlock structs.PointerBlock
	buffer := bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err != nil {
		return nil, err
	}

	return &pointerBlock, nil
}
