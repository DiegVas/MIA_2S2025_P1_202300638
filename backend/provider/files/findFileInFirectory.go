package files

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func findFileInDirectory(file *os.File, name string, dirInode structs.Inode, blockPos int32) (int32, error) {
	fmt.Printf("DEBUG FIND_DIR: Buscando '%s' en directorio (tamaño: %d)\n", name, dirInode.I_size)

	// Calcular cuántos bloques necesita este directorio basándose en su tamaño
	if dirInode.I_size <= 0 {
		fmt.Printf("DEBUG FIND_DIR: Directorio vacío (tamaño <= 0)\n")
		return -1, fmt.Errorf("directorio vacío")
	}

	blocksNeeded := (dirInode.I_size + structs.BLOCK_SIZE - 1) / structs.BLOCK_SIZE
	fmt.Printf("DEBUG FIND_DIR: Directorio necesita %d bloques\n", blocksNeeded)

	// Buscar en todos los bloques necesarios
	for i := int32(0); i < blocksNeeded && i < 12; i++ { // Máximo 12 bloques directos
		blockNum := dirInode.I_block[i]
		fmt.Printf("DEBUG FIND_DIR: Revisando bloque %d (I_block[%d])\n", blockNum, i)

		blockBytes := make([]byte, structs.BLOCK_SIZE)
		_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			fmt.Printf("DEBUG FIND_DIR: Error al leer bloque %d: %v\n", blockNum, err)
			continue
		}

		// Deserializar el bloque de directorio
		var folderBlock structs.FolderBlock
		buffer := bytes.NewReader(blockBytes)
		binary.Read(buffer, binary.LittleEndian, &folderBlock)

		fmt.Printf("DEBUG FIND_DIR: Contenido del bloque %d:\n", blockNum)

		// Buscar el archivo en las entradas de este bloque
		for j, content := range folderBlock.B_content {
			entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
			fmt.Printf("DEBUG FIND_DIR: Entrada %d: '%s' -> inodo %d\n", j, entryName, content.B_inodo)

			if entryName == name {
				fmt.Printf("DEBUG FIND_DIR: ¡Archivo encontrado! Retornando inodo %d\n", content.B_inodo)
				return content.B_inodo, nil
			}
		}
	}

	fmt.Printf("DEBUG FIND_DIR: Archivo '%s' no encontrado en ningún bloque\n", name)
	return -1, fmt.Errorf("archivo o directorio '%s' no encontrado", name)
}
