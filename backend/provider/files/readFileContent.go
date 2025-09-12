package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

func ReadFileContent(filePath string) (string, error) {
	// Verificar que haya una sesión activa
	session := structs.GetSession()
	if !session.Active {
		return "", errors.New("debe iniciar sesión para usar el comando CAT")
	} // Obtener la partición montada
	mountManager := mounted.GetMountManager()
	partition, partitionFound := mountManager.GetMountedPartition(session.ID)
	if !partitionFound {
		return "", errors.New("no hay ninguna partición montada para esta sesión")
	}

	// Leer el MBR del disco
	mbr, err := utils.ReadMBR(partition.Path)
	if err != nil {
		return "", fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	// Buscar la partición montada
	var targetPartition structs.PARTITION
	var found bool
	for _, p := range mbr.Partitions {
		partName := strings.TrimRight(string(p.Part_name[:]), "\x00")
		if partName == partition.Name {
			targetPartition = p
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("la partición '%s' no se encontró en el disco", partition.Name)
	}

	// Calcular las posiciones de las estructuras del sistema de archivos
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)

	// Abrir el archivo del disco
	file, err := os.Open(partition.Path)
	if err != nil {
		return "", fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	// Navegar por el path y encontrar el archivo
	inodeNum, err := findFileInodeByPath(file, filePath, inodePos, blockPos)
	if err != nil {
		return "", err
	}

	// Leer el inodo del archivo
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err = file.ReadAt(inodeBytes, int64(inodePos+(inodeNum*structs.INODE_SIZE)))
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo del archivo: %v", err)
	}

	// Deserializar el inodo
	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &inode)

	// Verificar que es un archivo (no un directorio)
	if inode.I_type != 0 {
		return "", errors.New("el path especificado no es un archivo")
	}

	// Verificar permisos de lectura
	currentUser, err := GetCurrentUserInfo(session, partition)
	if err != nil {
		return "", fmt.Errorf("error al obtener información del usuario: %v", err)
	}

	if !hasReadPermission(inode, currentUser) {
		return "", errors.New("no tiene permisos de lectura para este archivo")
	}

	// Leer el contenido del archivo
	content, err := readFileData(file, inode, blockPos)
	if err != nil {
		return "", fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	return content, nil
}
