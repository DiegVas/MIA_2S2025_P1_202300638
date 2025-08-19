package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
	"time"
)

func FindOrCreateParentDirectory(file *os.File, dirPath string, session *structs.Session,
	inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32, createParents bool) (int32, error) {

	if dirPath == "" {
		return 0, nil
	}

	components := strings.Split(dirPath, "/")
	currentInodeNum := int32(0)

	for _, component := range components {
		if component == "" {
			continue
		}

		// ? Buscar el directorio en el inodo actual
		nextInodeNum, exists := FindFileInDirectoryByInode(file, component, currentInodeNum, inodePos, blockPos)

		if !exists {
			if !createParents {
				return -1, fmt.Errorf("el directorio '%s' no existe", component)
			}

			//  ? Crear el directorio
			currentUser, err := GetCurrentUserInfo(session, mounted.MountedPartition{})
			if err != nil {
				return -1, fmt.Errorf("error al obtener información del usuario: %v", err)
			}

			nextInodeNum, err = CreateNewDirectory(file, component, currentUser, currentInodeNum,
				inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount)
			if err != nil {
				return -1, fmt.Errorf("error al crear directorio '%s': %v", component, err)
			}
		}

		currentInodeNum = nextInodeNum
	}

	return currentInodeNum, nil
}

func FindFileInDirectoryByInode(file *os.File, name string, dirInodeNum, inodePos, blockPos int32) (int32, bool) {
	// ? Leer el inodo del directorio
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err := file.ReadAt(inodeBytes, int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return -1, false
	}

	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &inode)

	// ? Leer el bloque del directorio
	blockNum := inode.I_block[0]
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	_, err = file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return -1, false
	}

	var folderBlock structs.FolderBlock
	buffer = bytes.NewReader(blockBytes)
	binary.Read(buffer, binary.LittleEndian, &folderBlock)

	// ? Buscar el archivo en las entradas del directorio
	for _, content := range folderBlock.B_content {
		entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if entryName == name {
			return content.B_inodo, true
		}
	}

	return -1, false
}

func CreateNewDirectory(file *os.File, dirName string, user *structs.User, parentInodeNum,
	inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32) (int32, error) {

	// Encontrar un inodo libre
	freeInodeNum, err := FindFreeInode(file, bmInodePos, inodeCount)
	if err != nil {
		return -1, err
	}

	// Encontrar un bloque libre
	freeBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, 1)
	if err != nil {
		return -1, err
	}
	blockNum := freeBlocks[0]

	// Crear el inodo del directorio
	now := time.Now()
	timeBytes := FormatTimeToBytes(now)

	dirInode := structs.Inode{
		I_uid:   user.UID,
		I_gid:   user.GID,
		I_size:  structs.FOLDER_SIZE,
		I_atime: timeBytes,
		I_ctime: timeBytes,
		I_mtime: timeBytes,
		I_block: [15]int32{blockNum},
		I_type:  1,   // 1 = directorio
		I_perm:  755, // rwxr-xr-x
	}

	// Escribir el inodo
	var inodeBuffer bytes.Buffer
	binary.Write(&inodeBuffer, binary.LittleEndian, dirInode)
	_, err = file.WriteAt(inodeBuffer.Bytes(), int64(inodePos+(freeInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return -1, err
	}

	// Crear el bloque del directorio con entradas . y ..
	folderBlock := structs.FolderBlock{}
	copy(folderBlock.B_content[0].B_name[:], ".")
	folderBlock.B_content[0].B_inodo = freeInodeNum
	copy(folderBlock.B_content[1].B_name[:], "..")
	folderBlock.B_content[1].B_inodo = parentInodeNum

	// Escribir el bloque del directorio
	var folderBuffer bytes.Buffer
	binary.Write(&folderBuffer, binary.LittleEndian, folderBlock)
	_, err = file.WriteAt(folderBuffer.Bytes(), int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return -1, err
	}

	// Marcar inodo y bloque como usados
	err = MarkInodeAsUsed(file, bmInodePos, freeInodeNum)
	if err != nil {
		return -1, err
	}

	err = MarkBlockAsUsed(file, bmBlockPos, blockNum)
	if err != nil {
		return -1, err
	}

	// Agregar la entrada al directorio padre
	err = AddFileToDirectory(file, dirName, freeInodeNum, parentInodeNum, inodePos, blockPos)
	if err != nil {
		return -1, err
	}

	return freeInodeNum, nil
}

func AddFileToDirectory(file *os.File, fileName string, fileInodeNum, dirInodeNum, inodePos, blockPos int32) error {
	// Leer el inodo del directorio
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err := file.ReadAt(inodeBytes, int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return err
	}

	var dirInode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &dirInode)

	// Leer el bloque del directorio
	blockNum := dirInode.I_block[0]
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	_, err = file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return err
	}

	var folderBlock structs.FolderBlock
	buffer = bytes.NewReader(blockBytes)
	binary.Read(buffer, binary.LittleEndian, &folderBlock)

	// Buscar una entrada libre
	for i := range folderBlock.B_content {
		entryName := strings.TrimRight(string(folderBlock.B_content[i].B_name[:]), "\x00")
		if entryName == "" {
			// Entrada libre encontrada
			copy(folderBlock.B_content[i].B_name[:], fileName)
			folderBlock.B_content[i].B_inodo = fileInodeNum

			// Escribir el bloque actualizado
			var folderBuffer bytes.Buffer
			binary.Write(&folderBuffer, binary.LittleEndian, folderBlock)
			_, err = file.WriteAt(folderBuffer.Bytes(), int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
			return err
		}
	}

	return errors.New("no hay espacio disponible en el directorio padre")
}
