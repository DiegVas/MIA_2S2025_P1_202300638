package files

import (
	"bytes"
	"encoding/binary"
	"main/structs"
	"os"
)

// DirectoryStats contiene estadísticas sobre un directorio
type DirectoryStats struct {
	TotalEntries    int32 // Total de entradas posibles
	UsedEntries     int32 // Entradas ocupadas
	FreeEntries     int32 // Entradas libres
	BlocksUsed      int32 // Bloques utilizados
	BlocksAvailable int32 // Bloques disponibles para expansión
	IsFull          bool  // Si el directorio está completamente lleno
}

// GetDirectoryStats analiza el estado de un directorio
func GetDirectoryStats(file *os.File, dirInodeNum, inodePos, blockPos int32) (*DirectoryStats, error) {
	// Leer el inodo del directorio
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err := file.ReadAt(inodeBytes, int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return nil, err
	}

	var dirInode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &dirInode)

	stats := &DirectoryStats{}

	// Analizar bloques directos (0-11)
	for i := 0; i < 12; i++ {
		if dirInode.I_block[i] != 0 {
			stats.BlocksUsed++
			entriesInBlock, err := countEntriesInBlock(file, dirInode.I_block[i], blockPos)
			if err != nil {
				continue
			}
			stats.UsedEntries += entriesInBlock
			stats.TotalEntries += 4 // Cada bloque tiene 4 entradas máximo
		} else {
			stats.BlocksAvailable++
		}
	}

	// TODO: Analizar bloques indirectos (slots 12, 13, 14)
	// Por ahora solo consideramos bloques directos

	stats.FreeEntries = stats.TotalEntries - stats.UsedEntries
	stats.IsFull = (stats.BlocksAvailable == 0 && stats.FreeEntries == 0)

	return stats, nil
}

// countEntriesInBlock cuenta las entradas ocupadas en un bloque de directorio
func countEntriesInBlock(file *os.File, blockNum, blockPos int32) (int32, error) {
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return 0, err
	}

	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockBytes)
	binary.Read(buffer, binary.LittleEndian, &folderBlock)

	var count int32
	for _, content := range folderBlock.B_content {
		// Verificar si la entrada tiene nombre (no está vacía)
		if content.B_name[0] != 0 {
			count++
		}
	}

	return count, nil
}

// CheckDirectoryCapacity verifica si un directorio puede aceptar más entradas
func CheckDirectoryCapacity(file *os.File, dirInodeNum, inodePos, blockPos int32) (bool, string, error) {
	stats, err := GetDirectoryStats(file, dirInodeNum, inodePos, blockPos)
	if err != nil {
		return false, "", err
	}

	if stats.IsFull {
		return false, "Directorio completamente lleno (48 entradas máximo con implementación actual)", nil
	}

	if stats.FreeEntries > 0 {
		return true, "Espacio disponible en bloques existentes", nil
	}

	if stats.BlocksAvailable > 0 {
		return true, "Se pueden crear nuevos bloques para el directorio", nil
	}

	return false, "Directorio en capacidad máxima", nil
}
