package files

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func readFileData(file *os.File, inode structs.Inode, blockPos int32) (string, error) {
	var content strings.Builder

	// Por simplicidad, solo manejaremos bloques directos
	for _, blockNum := range inode.I_block {
		if blockNum == 0 {
			break // No hay más bloques
		}

		// Leer el bloque de archivo
		blockBytes := make([]byte, structs.BLOCK_SIZE)
		_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque de datos: %v", err)
		}

		// Deserializar el bloque de archivo
		var fileBlock structs.FileBlock
		buffer := bytes.NewReader(blockBytes)
		binary.Read(buffer, binary.LittleEndian, &fileBlock)

		// Agregar el contenido (hasta encontrar el final o el tamaño del archivo)
		blockContent := string(fileBlock.B_content[:])
		// Remover caracteres nulos del final
		blockContent = strings.TrimRight(blockContent, "\x00")
		content.WriteString(blockContent)
	}

	result := content.String()

	// Truncar al tamaño del archivo
	if len(result) > int(inode.I_size) {
		result = result[:inode.I_size]
	}

	return result, nil
}
