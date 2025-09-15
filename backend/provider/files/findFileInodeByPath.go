package files

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func findFileInodeByPath(file *os.File, path string, inodePos, blockPos int32) (int32, error) {
	fmt.Printf("DEBUG FIND_PATH: Iniciando búsqueda para path: '%s'\n", path)

	// Si el path es la raíz, retornar el inodo 0 directamente
	if path == "/" {
		fmt.Printf("DEBUG FIND_PATH: Path es raíz, retornando inodo 0\n")
		return 0, nil
	}

	// Normalizar el path
	originalPath := path
	path = strings.TrimPrefix(path, "/")
	fmt.Printf("DEBUG FIND_PATH: Path original: '%s', normalizado: '%s'\n", originalPath, path)

	if path == "" {
		fmt.Printf("DEBUG FIND_PATH: Path normalizado vacío, retornando inodo 0 (raíz)\n")
		return 0, nil // Después de normalizar, si queda vacío, es la raíz
	}

	// Dividir el path en componentes
	components := strings.Split(path, "/")
	currentInodeNum := int32(0) // Comenzar desde el directorio raíz	// Para cada componente del path, navegar por el sistema de archivos
	for i, component := range components {
		fmt.Printf("DEBUG FIND: Procesando componente %d: '%s'\n", i, component)

		// Leer el inodo actual
		inodeBytes := make([]byte, structs.INODE_SIZE)
		_, err := file.ReadAt(inodeBytes, int64(inodePos+(currentInodeNum*structs.INODE_SIZE)))
		if err != nil {
			return -1, fmt.Errorf("error al leer el inodo: %v", err)
		}

		var inode structs.Inode
		buffer := bytes.NewReader(inodeBytes)
		binary.Read(buffer, binary.LittleEndian, &inode)

		if i == len(components)-1 {
			// El inodo actual debe ser un directorio para buscar el archivo
			if inode.I_type != 1 {
				return -1, fmt.Errorf("'%s' no es un directorio, no se puede buscar el archivo", strings.Join(components[:i], "/"))
			}
			// Buscar el archivo en el directorio
			return findFileInDirectory(file, component, inode, blockPos)
		} else {
			// Para componentes intermedios, el inodo actual debe ser un directorio
			if inode.I_type != 1 {
				return -1, fmt.Errorf("'%s' no es un directorio (I_type=%d)", component, inode.I_type)
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
