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
	"time"
)

func FindOrCreateParentDirectory(file *os.File, dirPath string, session *structs.Session,
	inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32, createParents bool) (int32, error) {

	if dirPath == "" {

		return 0, nil
	}

	components := strings.Split(dirPath, "/")
	currentInodeNum := int32(0)

	for i, component := range components {
		if component == "" {

			continue
		}

		fmt.Printf("DEBUG FIND_OR_CREATE: Procesando componente %d: '%s', inodo actual: %d\n", i, component, currentInodeNum)

		// ? Buscar el directorio en el inodo actual
		nextInodeNum, exists := FindFileInDirectoryByInode(file, component, currentInodeNum, inodePos, blockPos)

		if !exists {

			if !createParents {

				return -1, fmt.Errorf("el directorio '%s' no existe", component)
			}

			//  ? Obtener la partición montada
			mountManager := mounted.GetMountManager()
			partition, found := mountManager.GetMountedPartition(session.ID)
			if !found {
				fmt.Printf("DEBUG FIND_OR_CREATE: Error: no se encontró partición con ID '%s'\n", session.ID)
				return -1, fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
			}

			//  ? Crear el directorio
			currentUser, err := GetCurrentUserInfo(session, partition)
			if err != nil {
				fmt.Printf("DEBUG FIND_OR_CREATE: Error obteniendo usuario: %v\n", err)
				return -1, fmt.Errorf("error al obtener información del usuario: %v", err)
			}

			nextInodeNum, err = CreateNewDirectory(file, component, currentUser, currentInodeNum,
				inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount)
			if err != nil {
				fmt.Printf("DEBUG FIND_OR_CREATE: Error creando directorio '%s': %v\n", component, err)
				return -1, fmt.Errorf("error al crear directorio '%s': %v", component, err)
			}

		} else {
			fmt.Printf("DEBUG FIND_OR_CREATE: Componente '%s' ya existe con inodo: %d\n", component, nextInodeNum)
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

	// ? Buscar en todos los bloques directos del directorio
	for i := 0; i < 12; i++ { // Los primeros 12 son bloques directos
		if inode.I_block[i] != 0 {
			blockNum := inode.I_block[i]

			blockBytes := make([]byte, structs.BLOCK_SIZE)
			_, err = file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
			if err != nil {

				continue
			}

			var folderBlock structs.FolderBlock
			buffer = bytes.NewReader(blockBytes)
			binary.Read(buffer, binary.LittleEndian, &folderBlock)

			// ? Buscar el archivo en las entradas de este bloque
			for _, content := range folderBlock.B_content {
				entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
				if entryName == name {
					return content.B_inodo, true
				}
			}
		}
	}

	return -1, false
}

func CreateNewDirectory(file *os.File, dirName string, user *structs.User, parentInodeNum,
	inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32) (int32, error) {

	// Encontrar un inodo libre
	freeInodeNum, err := FindFreeInode(file, bmInodePos, inodeCount)
	if err != nil {
		fmt.Printf("❌ ERROR buscando inodo libre: %v\n", err)
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
	err = AddFileToDirectoryWithIndirect(file, dirName, freeInodeNum, parentInodeNum, inodePos, blockPos, bmBlockPos, blockCount)
	if err != nil {
		return -1, err
	}

	return freeInodeNum, nil
}

func AddFileToDirectory(file *os.File, fileName string, fileInodeNum, dirInodeNum, inodePos, blockPos, bmBlockPos, blockCount int32) error {
	// Usar la nueva implementación con soporte para bloques indirectos
	// Esta función mantiene compatibilidad con código existente
	return AddFileToDirectoryWithIndirect(file, fileName, fileInodeNum, dirInodeNum, inodePos, blockPos, bmBlockPos, blockCount)
}

// AddFileToDirectoryLegacy es la implementación original (solo bloques directos)
// Se mantiene para referencia y casos especiales
func AddFileToDirectoryLegacy(file *os.File, fileName string, fileInodeNum, dirInodeNum, inodePos, blockPos, bmBlockPos, blockCount int32) error {
	// Leer el inodo del directorio
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err := file.ReadAt(inodeBytes, int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return err
	}

	var dirInode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &dirInode)

	// Buscar en todos los bloques directos existentes
	for i := 0; i < 12; i++ { // Los primeros 12 son bloques directos
		if dirInode.I_block[i] != 0 {
			// Verificar si hay espacio en este bloque
			blockNum := dirInode.I_block[i]
			blockBytes := make([]byte, structs.BLOCK_SIZE)
			_, err = file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
			if err != nil {
				continue
			}

			var folderBlock structs.FolderBlock
			buffer = bytes.NewReader(blockBytes)
			binary.Read(buffer, binary.LittleEndian, &folderBlock)

			// Buscar una entrada libre en este bloque
			for j := range folderBlock.B_content {
				entryName := strings.TrimRight(string(folderBlock.B_content[j].B_name[:]), "\x00")
				if entryName == "" {
					// Entrada libre encontrada
					copy(folderBlock.B_content[j].B_name[:], fileName)
					folderBlock.B_content[j].B_inodo = fileInodeNum

					// Escribir el bloque actualizado
					var folderBuffer bytes.Buffer
					binary.Write(&folderBuffer, binary.LittleEndian, folderBlock)
					_, err = file.WriteAt(folderBuffer.Bytes(), int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
					return err
				}
			}
		} else {
			// Este slot está libre, crear un nuevo bloque de carpeta
			newBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, 1)
			if err != nil {
				// Si no hay bloques libres, retornar error específico
				return fmt.Errorf("no hay bloques libres disponibles para expandir el directorio: %v", err)
			}

			newBlockNum := newBlocks[0]
			dirInode.I_block[i] = newBlockNum

			// Crear nuevo FolderBlock
			var newFolderBlock structs.FolderBlock
			copy(newFolderBlock.B_content[0].B_name[:], fileName)
			newFolderBlock.B_content[0].B_inodo = fileInodeNum

			// Escribir el nuevo bloque
			var folderBuffer bytes.Buffer
			binary.Write(&folderBuffer, binary.LittleEndian, newFolderBlock)
			_, err = file.WriteAt(folderBuffer.Bytes(), int64(blockPos+(newBlockNum*structs.BLOCK_SIZE)))
			if err != nil {
				return err
			}

			// Marcar el bloque como usado
			err = MarkBlockAsUsed(file, bmBlockPos, newBlockNum)
			if err != nil {
				return err
			}

			// Actualizar el inodo del directorio
			var inodeBuffer bytes.Buffer
			binary.Write(&inodeBuffer, binary.LittleEndian, dirInode)
			_, err = file.WriteAt(inodeBuffer.Bytes(), int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
			return err
		}
	}

	return errors.New("no hay espacio disponible en el directorio padre - todos los bloques directos están llenos")
}

// CreateDirectory crea un directorio usando el comando mkdir
func CreateDirectory(params *structs.MkdirParams) error {
	session := structs.GetSession()
	if !session.Active {
		return errors.New("debe iniciar sesión para usar el comando MKDIR")
	}

	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return errors.New("no hay ninguna partición montada para esta sesión")
	}

	mbr, err := utils.ReadMBR(partition.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	// Buscar la partición montada
	var targetPartition structs.PARTITION
	var partitionFound bool
	for _, p := range mbr.Partitions {
		partName := strings.TrimRight(string(p.Part_name[:]), "\x00")
		if partName == partition.Name {
			targetPartition = p
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		return fmt.Errorf("la partición '%s' no se encontró en el disco", partition.Name)
	}

	// Calcular las posiciones de las estructuras del sistema de archivos
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)

	// Abrir el archivo del disco para escritura
	diskFile, err := os.OpenFile(partition.Path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	// Si el parámetro -p está presente, crear directorios padre
	if params.P {
		_, err = FindOrCreateParentDirectory(diskFile, params.Path, session,
			inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount, true)
		if err != nil {
			return fmt.Errorf("error al crear directorios padre: %v", err)
		}
		return nil
	}

	// Sin -p, verificar que el directorio padre existe y crear solo el directorio especificado
	parentPath := getParentPath(params.Path)
	dirName := getLastComponent(params.Path)

	// Buscar el inodo del directorio padre
	parentInodeNum, err := findFileInodeByPath(diskFile, parentPath, inodePos, blockPos)
	if err != nil {
		fmt.Printf("DEBUG MKDIR: Error buscando directorio padre: %v\n", err)
		return fmt.Errorf("el directorio padre no existe: %v", err)
	}

	fmt.Printf("DEBUG MKDIR: Directorio padre encontrado, inodo: %d\n", parentInodeNum)

	// Crear solo el directorio especificado
	currentUser, err := GetCurrentUserInfo(session, partition)
	if err != nil {
		return fmt.Errorf("error al obtener información del usuario: %v", err)
	}

	_, err = CreateNewDirectory(diskFile, dirName, currentUser, parentInodeNum,
		inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %v", err)
	}

	return nil
}

// getParentPath obtiene la ruta del directorio padre
func getParentPath(path string) string {

	if path == "/" {

		return "/"
	}

	path = strings.TrimSuffix(path, "/")

	lastSlash := strings.LastIndex(path, "/")

	if lastSlash == 0 {

		return "/"
	}
	if lastSlash == -1 {

		return "/"
	}

	result := path[:lastSlash]

	return result
}

// getLastComponent extrae el último componente de una ruta
func getLastComponent(path string) string {

	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		fmt.Printf("DEBUG getLastComponent: No hay parts, retornando ''\n")
		return ""
	}

	result := parts[len(parts)-1]
	return result
}
