package Analyzer

import (
	"bytes"
	"errors"
	"fmt"
	"main/structs"
	"main/utils"
	"os"
	"strconv"
	"strings"
)

// CreateGroup crea un nuevo grupo en la partición montada
func CreateGroup(params *GroupParams) error {
	// Verificar la sesión activa
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	
	if session.User != "root" {
		return errors.New("solo el usuario root puede crear grupos")
	}
	
	// Obtener la partición montada de la sesión actual
	mountManager := structs.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}
	
	// Leer el contenido actual de users.txt
	usersFile, err := readUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}
	
	// Verificar si el grupo ya existe
	lines := strings.Split(usersFile, "\n")
	maxGID := 1 // GID mínimo (root es 1)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		
		// Si es una línea de grupo (G)
		if len(fields) >= 3 && fields[1] == "G" {
			// Verificar si el grupo existe y no está eliminado (GID > 0)
			gid, _ := strconv.Atoi(fields[0])
			if gid > 0 && fields[2] == params.Name {
				return fmt.Errorf("el grupo '%s' ya existe", params.Name)
			}
			if gid > maxGID {
				maxGID = gid
			}
		}
	}
	
	// Crear el nuevo grupo con el siguiente GID
	newGID := maxGID + 1
	newGroupLine := fmt.Sprintf("%d, G, %s", newGID, params.Name)
	
	// Agregar la nueva línea al archivo
	if !strings.HasSuffix(usersFile, "\n") {
		usersFile += "\n"
	}
	usersFile += newGroupLine + "\n"
	
	// Escribir el archivo actualizado
	err = writeUsersFile(partition.Path, partition.Name, usersFile)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}
	
	return nil
}

// RemoveGroup elimina un grupo existente de la partición montada
func RemoveGroup(params *GroupParams) error {
	// Verificar la sesión activa
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	
	if session.User != "root" {
		return errors.New("solo el usuario root puede eliminar grupos")
	}
	
	// Obtener la partición montada de la sesión actual
	mountManager := structs.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}
	
	// Leer el contenido actual de users.txt
	usersFile, err := readUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}
	
	// Buscar y marcar el grupo como eliminado
	lines := strings.Split(usersFile, "\n")
	groupFound := false
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		fields := strings.Split(line, ",")
		for j := range fields {
			fields[j] = strings.TrimSpace(fields[j])
		}
		
		// Si es una línea de grupo (G) y coincide el nombre
		if len(fields) >= 3 && fields[1] == "G" && fields[2] == params.Name {
			gid, _ := strconv.Atoi(fields[0])
			if gid > 0 {
				// Verificar que no sea el grupo root
				if params.Name == "root" {
					return errors.New("no se puede eliminar el grupo root")
				}
				
				// Marcar como eliminado (GID = 0)
				lines[i] = fmt.Sprintf("0, G, %s", params.Name)
				groupFound = true
				break
			}
		}
	}
	
	if !groupFound {
		return fmt.Errorf("el grupo '%s' no existe o ya fue eliminado", params.Name)
	}
	
	// Reconstruir el archivo
	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	
	// Escribir el archivo actualizado
	err = writeUsersFile(partition.Path, partition.Name, newContent)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}
	
	return nil
}

// readUsersFile lee el contenido del archivo users.txt de la partición
func readUsersFile(diskPath, partitionName string) (string, error) {
	// Leer el MBR del disco
	mbr, err := utils.ReadMBR(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al leer el MBR del disco: %v", err)
	}
	
	// Buscar la partición
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
	
	// Calcular la posición del archivo users.txt (asumimos que está en el bloque 1)
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE
	
	// Posiciones de las estructuras del sistema de archivos
	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)
	
	// Posición del bloque de datos de users.txt (asumimos que es el bloque 1)
	usersBlockPos := blockPos + structs.BLOCK_SIZE // Bloque 1 (el bloque 0 es para el directorio raíz)
	
	// Leer el bloque de datos
	file, err := os.Open(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()
	
	// Leer el bloque de datos de users.txt
	block := make([]byte, structs.BLOCK_SIZE)
	_, err = file.ReadAt(block, int64(usersBlockPos))
	if err != nil {
		return "", fmt.Errorf("error al leer el bloque de datos: %v", err)
	}
	
	// Convertir a string y eliminar caracteres nulos
	content := string(bytes.Trim(block, "\x00"))
	
	return content, nil
}

// writeUsersFile escribe el contenido actualizado en el archivo users.txt de la partición
func writeUsersFile(diskPath, partitionName, content string) error {
	// Leer el MBR del disco
	mbr, err := utils.ReadMBR(diskPath)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}
	
	// Buscar la partición
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
	
	// Calcular la posición del archivo users.txt (asumimos que está en el bloque 1)
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE
	
	// Posiciones de las estructuras del sistema de archivos
	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)
	
	// Posición del bloque de datos de users.txt (asumimos que es el bloque 1)
	usersBlockPos := blockPos + structs.BLOCK_SIZE // Bloque 1 (el bloque 0 es para el directorio raíz)
	
	// Obtener el inodo de users.txt para actualizar su tamaño
	usersInodePos := inodePos + structs.INODE_SIZE // Inodo 1 (el inodo 0 es para el directorio raíz)
	
	// Abrir el archivo del disco para escritura
	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()
	
	// Leer el inodo de users.txt
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err = file.ReadAt(inodeBytes, int64(usersInodePos))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}
	
	// Deserializar el inodo
	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	
	// Leer cada campo del inodo
	// (Nota: esto es una simplificación, idealmente deberíamos usar binary.Read para cada campo)
	buffer.Seek(0, 0)
	
	// Actualizar el tamaño del inodo
	inode.I_size = int32(len(content))
	
	// Escribir el tamaño actualizado en el inodo (offset 8 para el tamaño)
	sizeBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		sizeBytes[i] = byte((inode.I_size >> (i * 8)) & 0xFF)
	}
	_, err = file.WriteAt(sizeBytes, int64(usersInodePos+8))
	if err != nil {
		return fmt.Errorf("error al actualizar el tamaño del inodo: %v", err)
	}
	
	// Preparar el bloque de datos con el nuevo contenido
	blockData := make([]byte, structs.BLOCK_SIZE)
	copy(blockData, content)
	
	// Escribir el bloque de datos actualizado
	_, err = file.WriteAt(blockData, int64(usersBlockPos))
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de datos: %v", err)
	}
	
	return nil
}
