package users

import (
	"bytes"
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

	// Calcular la posición del archivo users.txt
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	// Posiciones de las estructuras del sistema de archivos
	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)

	// Posición del bloque de datos de users.txt )
	usersBlockPos := blockPos + structs.BLOCK_SIZE

	// ? Leer el bloque de datos
	file, err := os.Open(diskPath)

	if err != nil {
		return "", fmt.Errorf("error al abrir el disco: %v", err)
	}

	defer file.Close()

	// ? Leer el bloque de datos de users.txt
	block := make([]byte, structs.BLOCK_SIZE)
	_, err = file.ReadAt(block, int64(usersBlockPos))
	if err != nil {
		return "", fmt.Errorf("error al leer el bloque de datos: %v", err)
	}

	// ! Convertir a string y eliminar caracteres nulos
	content := string(bytes.Trim(block, "\x00"))

	return content, nil
}
