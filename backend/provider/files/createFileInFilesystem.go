package files

import (
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
)

func createFileInFilesystem(file *os.File, path, content string, size int32, session *structs.Session,
	inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32, createParents bool) error {

	// ? Normalizar el path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return errors.New("path del archivo no puede estar vacío")
	}

	// ? Dividir el path en directorio padre y nombre del archivo
	pathComponents := strings.Split(path, "/")
	fileName := pathComponents[len(pathComponents)-1]
	dirPath := strings.Join(pathComponents[:len(pathComponents)-1], "/")

	//  ? Encontrar o crear el directorio padre
	parentInodeNum, err := FindOrCreateParentDirectory(file, dirPath, session,
		inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount, createParents)
	if err != nil {
		return err
	}

	// ? Verificar permisos de escritura en el directorio padre
	if !HasWritePermissionOnDirectory(file, parentInodeNum, session, inodePos) {
		return errors.New("no tiene permisos de escritura en el directorio padre")
	}

	// ! Verificar si el archivo ya existe
	_, exists := FindFileInDirectoryByInode(file, fileName, parentInodeNum, inodePos, blockPos)
	if exists {
		return fmt.Errorf("el archivo '%s' ya existe", fileName)
	}

	// ! Obtener la partición montada
	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}

	// ! Obtener información del usuario actual para el archivo
	currentUser, err := GetCurrentUserInfo(session, partition)
	if err != nil {
		return fmt.Errorf("error al obtener información del usuario: %v", err)
	}

	// ? Crear el archivo
	err = CreateNewFile(file, fileName, content, size, currentUser, parentInodeNum,
		inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount)
	if err != nil {
		return err
	}

	return nil
}
