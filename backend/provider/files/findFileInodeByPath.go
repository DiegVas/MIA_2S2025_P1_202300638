package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func findFileInodeByPath(file *os.File, path string, inodePos, blockPos int32) (int32, error) {
	// Normalizar el path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return 0, errors.New("path vacío")
	}

	// Dividir el path en componentes
	components := strings.Split(path, "/")
	currentInodeNum := int32(0) // Comenzar desde el directorio raíz

	// Para cada componente del path, navegar por el sistema de archivos
	for i, component := range components {
		// Leer el inodo actual
		inodeBytes := make([]byte, structs.INODE_SIZE)
		_, err := file.ReadAt(inodeBytes, int64(inodePos+(currentInodeNum*structs.INODE_SIZE)))
		if err != nil {
			return -1, fmt.Errorf("error al leer el inodo: %v", err)
		}

		var inode structs.Inode
		buffer := bytes.NewReader(inodeBytes)
		binary.Read(buffer, binary.LittleEndian, &inode)

		// Si es el último componente, debe ser un archivo
		if i == len(components)-1 {
			if inode.I_type != 1 { // No es un directorio
				return currentInodeNum, nil
			}
			// Buscar el archivo en el directorio
			return findFileInDirectory(file, component, inode, blockPos)
		} else {
			// Debe ser un directorio
			if inode.I_type != 1 {
				return -1, fmt.Errorf("'%s' no es un directorio", component)
			}
			// Buscar el siguiente directorio
			nextInodeNum, err := findFileInDirectory(file, component, inode, blockPos)
			if err != nil {
				return -1, err
			}
			currentInodeNum = nextInodeNum
		}
	}

	return currentInodeNum, nil
}
