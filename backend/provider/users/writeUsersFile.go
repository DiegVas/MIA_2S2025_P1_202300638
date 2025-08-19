package users

import (
	"bytes"
	"fmt"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

func WriteUsersFile(diskPath, partitionName, content string) error {

	// ? Leer el MBR del disco
	mbr, err := utils.ReadMBR(diskPath)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
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
		return fmt.Errorf("la partición '%s' no se encontró en el disco", partitionName)
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

	// Posición del bloque de datos de users.txt
	usersBlockPos := blockPos + structs.BLOCK_SIZE

	// Obtener el inodo de users.txt para actualizar su tamaño
	usersInodePos := inodePos + structs.INODE_SIZE

	// ? Abrir el archivo del disco para escritura
	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	// ? Leer el inodo de users.txt
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err = file.ReadAt(inodeBytes, int64(usersInodePos))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)

	buffer.Seek(0, 0)

	// ? Actualizar el tamaño del inodo
	inode.I_size = int32(len(content))

	// ? Escribir el tamaño actualizado en el inodo (offset 8 para el tamaño)
	sizeBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		sizeBytes[i] = byte((inode.I_size >> (i * 8)) & 0xFF)
	}
	_, err = file.WriteAt(sizeBytes, int64(usersInodePos+8))
	if err != nil {
		return fmt.Errorf("error al actualizar el tamaño del inodo: %v", err)
	}

	// ? Preparar el bloque de datos con el nuevo contenido
	blockData := make([]byte, structs.BLOCK_SIZE)
	copy(blockData, content)

	// ? Escribir el bloque de datos actualizado
	_, err = file.WriteAt(blockData, int64(usersBlockPos))
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de datos: %v", err)
	}

	return nil
}
