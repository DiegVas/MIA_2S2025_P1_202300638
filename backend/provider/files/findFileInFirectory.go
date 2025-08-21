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
	// Buscar en todos los bloques directos del directorio
	for i := 0; i < 12; i++ { // Los primeros 12 son bloques directos
		if dirInode.I_block[i] != 0 {
			blockNum := dirInode.I_block[i]
			blockBytes := make([]byte, structs.BLOCK_SIZE)
			_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
			if err != nil {
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
		}
	}

	return -1, fmt.Errorf("archivo o directorio '%s' no encontrado", name)
}
