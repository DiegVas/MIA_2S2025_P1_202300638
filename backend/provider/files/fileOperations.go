package files

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"time"
)

func CreateNewFile(file *os.File, fileName, content string, size int32, user *structs.User,
	parentInodeNum, inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32) error {

	// ! Encontrar un inodo libre
	freeInodeNum, err := FindFreeInode(file, bmInodePos, inodeCount)
	if err != nil {
		return err
	}

	// ? Encontrar bloques libres para el contenido
	blocksNeeded := (size + structs.BLOCK_SIZE - 1) / structs.BLOCK_SIZE
	if blocksNeeded == 0 {
		blocksNeeded = 1 // Al menos un bloque para archivos vacíos
	}

	freeBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, blocksNeeded)
	if err != nil {
		return err
	}

	// ? Crear el inodo del archivo
	now := time.Now()
	timeBytes := FormatTimeToBytes(now)

	fileInode := structs.Inode{
		I_uid:   user.UID,
		I_gid:   user.GID,
		I_size:  size,
		I_atime: timeBytes,
		I_ctime: timeBytes,
		I_mtime: timeBytes,
		I_type:  0,
		I_perm:  664,
	}

	// ! Asignar bloques al inodo
	for i, blockNum := range freeBlocks {
		if i < 15 {
			fileInode.I_block[i] = blockNum
		}
	}

	// ? Escribir el inodo
	var inodeBuffer bytes.Buffer
	binary.Write(&inodeBuffer, binary.LittleEndian, fileInode)
	_, err = file.WriteAt(inodeBuffer.Bytes(), int64(inodePos+(freeInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo del archivo: %v", err)
	}

	// ? Escribir el contenido en los bloques
	err = WriteContentToBlocks(file, content, freeBlocks, blockPos)
	if err != nil {
		return err
	}

	// ? Marcar el inodo como usado en el bitmap
	err = MarkInodeAsUsed(file, bmInodePos, freeInodeNum)
	if err != nil {
		return err
	}

	// ?  Marcar los bloques como usados en el bitmap
	for _, blockNum := range freeBlocks {
		err = MarkBlockAsUsed(file, bmBlockPos, blockNum)
		if err != nil {
			return err
		}
	}

	// ? Agregar la entrada al directorio padre
	err = AddFileToDirectory(file, fileName, freeInodeNum, parentInodeNum, inodePos, blockPos)
	if err != nil {
		return err
	}

	return nil
}

func FormatTimeToBytes(t time.Time) [16]byte {
	var timeBytes [16]byte
	timeStr := t.Format("2006-01-02 15:04")
	copy(timeBytes[:], timeStr)
	return timeBytes
}
