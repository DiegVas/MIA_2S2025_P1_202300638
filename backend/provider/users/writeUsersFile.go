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

	// Deserializar el inodo completo
	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &inode)

	// ? Actualizar el tamaño del inodo
	inode.I_size = int32(len(content))

	// Calcular cuántos bloques necesitamos
	blocksNeeded := (len(content) + structs.FILE_CONTENT_SIZE - 1) / structs.FILE_CONTENT_SIZE

	// Asegurar que tenemos suficientes bloques asignados
	err = ensureUsersFileBlocks(file, &inode, blocksNeeded, bmBlockPos, blockPos, blockCount)
	if err != nil {
		return fmt.Errorf("error al asignar bloques: %v", err)
	}

	// Escribir el contenido usando múltiples bloques
	err = writeContentToUsersBlocks(file, content, &inode, blockPos)
	if err != nil {
		return fmt.Errorf("error al escribir contenido: %v", err)
	}

	// Serializar y escribir el inodo completo actualizado
	var inodeBuffer bytes.Buffer
	binary.Write(&inodeBuffer, binary.LittleEndian, inode)
	_, err = file.WriteAt(inodeBuffer.Bytes(), int64(usersInodePos))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo actualizado: %v", err)
	}

	return nil
}

// ensureUsersFileBlocks asegura que el inodo tenga suficientes bloques asignados
func ensureUsersFileBlocks(file *os.File, inode *structs.Inode, blocksNeeded int, bmBlockPos, _, blockCount int32) error {

	// Soportamos hasta 12 bloques directos (suficiente para un archivo grande)
	if blocksNeeded > 12 {
		return fmt.Errorf("archivo demasiado grande, máximo 12 bloques directos soportados")
	}

	// Si ya tenemos suficientes bloques, no hacer nada
	currentBlocks := 0
	for i := 0; i < 12; i++ {
		if inode.I_block[i] != 0 {
			currentBlocks++
		} else {
			break
		}
	}

	if currentBlocks >= blocksNeeded {
		return nil // Ya tenemos suficientes bloques
	}

	// Necesitamos asignar más bloques
	for i := currentBlocks; i < blocksNeeded && i < 12; i++ {
		newBlockNum, err := findFreeBlock(file, bmBlockPos, blockCount)
		if err != nil {
			return fmt.Errorf("error al encontrar bloque libre: %v", err)
		}

		inode.I_block[i] = newBlockNum

		// Marcar el bloque como ocupado en el bitmap
		err = markBlockAsUsed(file, bmBlockPos, newBlockNum)
		if err != nil {
			return fmt.Errorf("error al marcar bloque como usado: %v", err)
		}
	}

	return nil
}

// writeContentToUsersBlocks escribe contenido usando múltiples bloques
func writeContentToUsersBlocks(file *os.File, content string, inode *structs.Inode, blockPos int32) error {
	contentBytes := []byte(content)
	writtenBytes := 0

	for i := 0; i < 12 && inode.I_block[i] != 0 && writtenBytes < len(contentBytes); i++ {
		blockData := make([]byte, structs.BLOCK_SIZE)

		start := writtenBytes
		end := start + structs.FILE_CONTENT_SIZE
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		if start < len(contentBytes) {
			copy(blockData, contentBytes[start:end])
		}

		blockNum := inode.I_block[i]
		position := int64(blockPos + (blockNum * structs.BLOCK_SIZE))

		_, err := file.WriteAt(blockData, position)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque %d: %v", blockNum, err)
		}

		writtenBytes = end
	}

	return nil
}

// findFreeBlock encuentra un bloque libre en el bitmap
func findFreeBlock(file *os.File, bmBlockPos, blockCount int32) (int32, error) {
	bitmap := make([]byte, blockCount)
	_, err := file.ReadAt(bitmap, int64(bmBlockPos))
	if err != nil {
		return -1, err
	}

	for i := int32(0); i < blockCount; i++ {
		if bitmap[i] == 0 {
			return i, nil
		}
	}

	return -1, fmt.Errorf("no hay bloques libres disponibles")
}

// markBlockAsUsed marca un bloque como usado en el bitmap
func markBlockAsUsed(file *os.File, bmBlockPos, blockNum int32) error {
	position := int64(bmBlockPos + blockNum)
	_, err := file.WriteAt([]byte{1}, position)
	return err
}
