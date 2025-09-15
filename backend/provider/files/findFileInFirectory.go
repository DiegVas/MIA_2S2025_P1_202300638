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
	// Calcular cuántos bloques necesita este directorio basándose en su tamaño
	if dirInode.I_size <= 0 {
		fmt.Printf("DEBUG FIND_DIR: Directorio vacío (tamaño <= 0)\n")
		return -1, fmt.Errorf("directorio vacío")
	}

	// Buscar en todos los bloques directos válidos (no solo los necesarios según tamaño)
	for i := int32(0); i < 12; i++ { // Revisar todos los 12 bloques directos posibles
		blockNum := dirInode.I_block[i]
		// Nota: El bloque 0 SÍ puede ser válido, solo omitimos si es -1 o similar
		if blockNum >= 0 { // Verificar que el bloque sea válido (>= 0)

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

			// Buscar el archivo en las entradas de este bloque
			for _, content := range folderBlock.B_content {
				entryName := strings.TrimRight(string(content.B_name[:]), "\x00")

				if entryName == name {
					return content.B_inodo, nil
				}
			}
		} // Cerrar el if blockNum >= 0
	}

	return -1, fmt.Errorf("archivo o directorio '%s' no encontrado", name)
}
