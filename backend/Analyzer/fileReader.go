package Analyzer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/structs"
	"main/utils"
	"os"
	"strconv"
	"strings"
)

// ReadFileContent lee el contenido de un archivo dado su path en el sistema de archivos
func ReadFileContent(filePath string) (string, error) {
	// Verificar que haya una sesión activa
	session := structs.GetSession()
	if !session.Active {
		return "", errors.New("debe iniciar sesión para usar el comando CAT")
	}

	// Obtener la partición montada
	mountManager := structs.GetMountManager()
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
	currentUser, err := getCurrentUserInfo(session, partition)
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

// findFileInodeByPath navega por el sistema de archivos para encontrar un archivo
func findFileInodeByPath(file *os.File, path string, inodePos, blockPos int32) (int32, error) {
	// Normalizar el path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return 0, errors.New("path vacío")
	}

	// Dividir el path en componentes
	components := strings.Split(path, "/")
	currentInodeNum := int32(0) // Comenzar desde el directorio raíz

	// Para cada componente del path, navegar por el sistema de archivos
	for i, component := range components {
		// Leer el inodo actual
		inodeBytes := make([]byte, structs.INODE_SIZE)
		_, err := file.ReadAt(inodeBytes, int64(inodePos+(currentInodeNum*structs.INODE_SIZE)))
		if err != nil {
			return -1, fmt.Errorf("error al leer el inodo: %v", err)
		}

		var inode structs.Inode
		buffer := bytes.NewReader(inodeBytes)
		binary.Read(buffer, binary.LittleEndian, &inode)

		// Si es el último componente, debe ser un archivo
		if i == len(components)-1 {
			if inode.I_type != 1 { // No es un directorio
				return currentInodeNum, nil
			}
			// Buscar el archivo en el directorio
			return findFileInDirectory(file, component, inode, blockPos)
		} else {
			// Debe ser un directorio
			if inode.I_type != 1 {
				return -1, fmt.Errorf("'%s' no es un directorio", component)
			}
			// Buscar el siguiente directorio
			nextInodeNum, err := findFileInDirectory(file, component, inode, blockPos)
			if err != nil {
				return -1, err
			}
			currentInodeNum = nextInodeNum
		}
	}

	return currentInodeNum, nil
}

// findFileInDirectory busca un archivo o directorio dentro de un directorio
func findFileInDirectory(file *os.File, name string, dirInode structs.Inode, blockPos int32) (int32, error) {
	// Leer el bloque del directorio
	blockNum := dirInode.I_block[0] // Usar el primer bloque directo
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return -1, fmt.Errorf("error al leer el bloque del directorio: %v", err)
	}

	// Deserializar el bloque de directorio
	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockBytes)
	binary.Read(buffer, binary.LittleEndian, &folderBlock)

	// Buscar el archivo en las entradas del directorio
	for _, content := range folderBlock.B_content {
		entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if entryName == name {
			return content.B_inodo, nil
		}
	}

	return -1, fmt.Errorf("archivo o directorio '%s' no encontrado", name)
}

// hasReadPermission verifica si el usuario actual tiene permisos de lectura
func hasReadPermission(inode structs.Inode, user *structs.User) bool {
	if user == nil {
		return false
	}

	// Si es root, tiene todos los permisos
	if user.UID == 1 {
		return true
	}

	// Obtener los permisos del archivo
	perms := inode.I_perm

	// Verificar permisos de propietario
	if inode.I_uid == user.UID {
		return (perms & 0400) != 0 // Permiso de lectura del propietario
	}

	// Verificar permisos de grupo
	if inode.I_gid == user.GID {
		return (perms & 0040) != 0 // Permiso de lectura del grupo
	}

	// Verificar permisos de otros
	return (perms & 0004) != 0 // Permiso de lectura de otros
}

// readFileData lee el contenido de un archivo desde sus bloques
func readFileData(file *os.File, inode structs.Inode, blockPos int32) (string, error) {
	var content strings.Builder
	
	// Por simplicidad, solo manejaremos bloques directos
	for _, blockNum := range inode.I_block {
		if blockNum == 0 {
			break // No hay más bloques
		}

		// Leer el bloque de archivo
		blockBytes := make([]byte, structs.BLOCK_SIZE)
		_, err := file.ReadAt(blockBytes, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque de datos: %v", err)
		}

		// Deserializar el bloque de archivo
		var fileBlock structs.FileBlock
		buffer := bytes.NewReader(blockBytes)
		binary.Read(buffer, binary.LittleEndian, &fileBlock)

		// Agregar el contenido (hasta encontrar el final o el tamaño del archivo)
		blockContent := string(fileBlock.B_content[:])
		// Remover caracteres nulos del final
		blockContent = strings.TrimRight(blockContent, "\x00")
		content.WriteString(blockContent)
	}

	result := content.String()
	
	// Truncar al tamaño del archivo
	if len(result) > int(inode.I_size) {
		result = result[:inode.I_size]
	}

	return result, nil
}

// getCurrentUserInfo obtiene la información del usuario actual desde la sesión
func getCurrentUserInfo(session *structs.Session, partition structs.MountedPartition) (*structs.User, error) {
	// Leer el archivo users.txt para obtener información del usuario
	usersContent, err := readUsersFile(partition.Path, partition.Name)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	lines := strings.Split(usersContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		// Formato: UID, Tipo, Grupo, Usuario, Contraseña
		if len(parts) >= 5 && strings.TrimSpace(parts[1]) == "U" {
			username := strings.TrimSpace(parts[3])
			if username == session.User {
				// Convertir UID y GID
				uid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				groupName := strings.TrimSpace(parts[2])
				
				// Buscar el GID del grupo
				gid := findGroupGID(usersContent, groupName)
				
				return &structs.User{
					Username: username,
					Password: strings.TrimSpace(parts[4]),
					Group:    groupName,
					UID:      int32(uid),
					GID:      int32(gid),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("usuario '%s' no encontrado", session.User)
}

// findGroupGID busca el GID de un grupo en el contenido de users.txt
func findGroupGID(usersContent, groupName string) int {
	lines := strings.Split(usersContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 && strings.TrimSpace(parts[1]) == "G" {
			if strings.TrimSpace(parts[2]) == groupName {
				gid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				return gid
			}
		}
	}
	return 0
}
