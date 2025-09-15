package users

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

func ReadUsersFile(diskPath, partitionName string) (string, error) {
	// ? Leer el MBR del disco
	mbr, err := utils.ReadMBR(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	// ? Buscar la partición
	var targetPartition structs.PARTITION
	found := false

	for _, p := range mbr.Partitions {
		partName := strings.TrimRight(string(p.Part_name[:]), "\x00")
		if partName == partitionName {
			targetPartition = p
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("la partición '%s' no se encontró en el disco", partitionName)
	}

	// Calcular las posiciones de las estructuras del sistema de archivos
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)

	// Posición del inodo de users.txt (inodo 1)
	usersInodePos := inodePos + structs.INODE_SIZE

	// ? Abrir el archivo del disco
	file, err := os.Open(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	// ? Leer el inodo de users.txt
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err = file.ReadAt(inodeBytes, int64(usersInodePos))
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Deserializar el inodo
	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &inode)

	// Leer el contenido usando múltiples bloques
	content, err := readUsersFileData(file, inode, blockPos)
	if err != nil {
		return "", fmt.Errorf("error al leer el contenido: %v", err)
	}

	return content, nil
}

// readUsersFileData lee el contenido del archivo users.txt usando múltiples bloques
func readUsersFileData(file *os.File, inode structs.Inode, blockPos int32) (string, error) {
	var content strings.Builder
	totalBytesRead := 0

	// Leer bloques directos (slots 0-11)
	for i := 0; i < 12 && inode.I_block[i] != 0 && totalBytesRead < int(inode.I_size); i++ {
		blockNum := inode.I_block[i]

		// Leer el bloque completo
		blockBytes := make([]byte, structs.BLOCK_SIZE)
		_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque %d: %v", blockNum, err)
		}

		// Determinar cuántos bytes leer de este bloque
		bytesToRead := structs.FILE_CONTENT_SIZE
		if totalBytesRead+bytesToRead > int(inode.I_size) {
			bytesToRead = int(inode.I_size) - totalBytesRead
		}

		// Agregar al contenido
		content.Write(blockBytes[:bytesToRead])
		totalBytesRead += bytesToRead
	}

	return content.String(), nil
}
