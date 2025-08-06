package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"time"
)

// FormatPartition formatea una partición en formato ext2
func FormatPartition(params *structs.MkfsParams) error {
	// Obtener la información de la partición montada
	mountManager := structs.GetMountManager()
	partition, found := mountManager.GetMountedPartition(params.ID)
	
	if !found {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s'", params.ID)
	}
	
	// Leer el MBR del disco
	mbr, err := ReadMBR(partition.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}
	
	// Buscar la partición dentro del MBR
	var targetPartition structs.PARTITION
	found = false
	
	for _, p := range mbr.Partitions {
		// Convertir el nombre de la partición a string (eliminando caracteres nulos)
		partName := string(bytes.Trim(p.Part_name[:], "\x00"))
		
		if partName == partition.Name {
			targetPartition = p
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("error: la partición '%s' no se encontró en el disco", partition.Name)
	}
	
	// Calcular el número de inodos y bloques
	// Un inodo ocupa 76 bytes y un bloque 64 bytes
	// Usaremos aproximadamente el 5% de la partición para inodos
	partitionSize := targetPartition.Part_size
	
	// Calcular el número de inodos (5% del tamaño)
	inodeCount := int32(float64(partitionSize) * 0.05 / float64(structs.INODE_SIZE))
	
	// El resto se usa para bloques
	blockCount := (partitionSize - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE
	
	// Crear el superbloque
	superblock := structs.Superblock{
		S_filesystem_type:   2, // ext2
		S_inodes_count:      inodeCount,
		S_blocks_count:      blockCount,
		S_free_blocks_count: blockCount - 2, // Un bloque para el directorio raíz y otro para users.txt
		S_free_inodes_count: inodeCount - 2, // Un inodo para el directorio raíz y otro para users.txt
		S_mtime:             getCurrentTimeBytes(),
		S_umtime:            [16]byte{}, // Inicialmente vacío
		S_mnt_count:         1,
		S_magic:             structs.EXT2_MAGIC,
		S_inode_size:        structs.INODE_SIZE,
		S_block_size:        structs.BLOCK_SIZE,
		S_first_ino:         2, // Los primeros inodos son para directorio raíz y users.txt
		S_first_blo:         2, // Los primeros bloques son para el directorio raíz y users.txt
	}
	
	// Calcular las posiciones de las estructuras
	startPosition := targetPartition.Part_start
	
	// Posición del superbloque
	superbPos := startPosition
	
	// Posición del bitmap de inodos
	bmInodePos := superbPos + structs.SUPERBLOCK_SIZE
	superblock.S_bm_inode_start = bmInodePos
	
	// Posición del bitmap de bloques
	bmBlockPos := bmInodePos + inodeCount
	superblock.S_bm_block_start = bmBlockPos
	
	// Posición de la tabla de inodos
	inodePos := bmBlockPos + blockCount
	superblock.S_inode_start = inodePos
	
	// Posición de la tabla de bloques
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)
	superblock.S_block_start = blockPos
	
	// Abrir el archivo del disco
	diskFile, err := os.OpenFile(partition.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()
	
	// Escribir el superbloque
	var superbBuffer bytes.Buffer
	binary.Write(&superbBuffer, binary.LittleEndian, superblock)
	_, err = diskFile.WriteAt(superbBuffer.Bytes(), int64(superbPos))
	if err != nil {
		return fmt.Errorf("error al escribir el superbloque: %v", err)
	}
	
	// Inicializar el bitmap de inodos (todos libres)
	inodeBitmap := make([]byte, inodeCount)
	for i := range inodeBitmap {
		inodeBitmap[i] = 0 // 0 = libre
	}
	// Marcar los primeros inodos como usados (raíz y users.txt)
	inodeBitmap[0] = 1 // Directorio raíz
	inodeBitmap[1] = 1 // Archivo users.txt
	
	// Escribir el bitmap de inodos
	_, err = diskFile.WriteAt(inodeBitmap, int64(bmInodePos))
	if err != nil {
		return fmt.Errorf("error al escribir el bitmap de inodos: %v", err)
	}
	
	// Inicializar el bitmap de bloques (todos libres)
	blockBitmap := make([]byte, blockCount)
	for i := range blockBitmap {
		blockBitmap[i] = 0 // 0 = libre
	}
	// Marcar los primeros bloques como usados (directorio raíz y users.txt)
	blockBitmap[0] = 1 // Directorio raíz
	blockBitmap[1] = 1 // Archivo users.txt
	
	// Escribir el bitmap de bloques
	_, err = diskFile.WriteAt(blockBitmap, int64(bmBlockPos))
	if err != nil {
		return fmt.Errorf("error al escribir el bitmap de bloques: %v", err)
	}
	
	// Crear el inodo para el directorio raíz
	rootInode := structs.Inode{
		I_uid:   0, // root
		I_gid:   0, // root
		I_size:  structs.FOLDER_SIZE,
		I_atime: getCurrentTimeBytes(),
		I_ctime: getCurrentTimeBytes(),
		I_mtime: getCurrentTimeBytes(),
		I_block: [15]int32{0}, // El primer bloque es el directorio raíz
		I_type:  1,            // 1 = directorio
		I_perm:  664,          // Permisos rw-rw-r--
	}
	
	// Escribir el inodo raíz
	var rootInodeBuffer bytes.Buffer
	binary.Write(&rootInodeBuffer, binary.LittleEndian, rootInode)
	_, err = diskFile.WriteAt(rootInodeBuffer.Bytes(), int64(inodePos))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo raíz: %v", err)
	}
	
	// Crear el inodo para users.txt
	usersInode := structs.Inode{
		I_uid:   1, // Usuario root (UID=1)
		I_gid:   1, // Grupo root (GID=1)
		I_size:  int32(len("1, G, root \n1, U, root, root, 123 \n")), // Tamaño del contenido inicial
		I_atime: getCurrentTimeBytes(),
		I_ctime: getCurrentTimeBytes(),
		I_mtime: getCurrentTimeBytes(),
		I_block: [15]int32{1}, // Bloque 1 para el contenido del archivo
		I_type:  0,           // 0 = archivo
		I_perm:  664,         // Permisos rw-rw-r--
	}
	
	// Escribir el inodo de users.txt
	var usersInodeBuffer bytes.Buffer
	binary.Write(&usersInodeBuffer, binary.LittleEndian, usersInode)
	_, err = diskFile.WriteAt(usersInodeBuffer.Bytes(), int64(inodePos+structs.INODE_SIZE))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo de users.txt: %v", err)
	}
	
	// Crear el bloque para el directorio raíz
	rootFolder := structs.FolderBlock{}
	
	// Entrada "."
	copy(rootFolder.B_content[0].B_name[:], ".")
	rootFolder.B_content[0].B_inodo = 0 // Apunta a sí mismo
	
	// Entrada ".."
	copy(rootFolder.B_content[1].B_name[:], "..")
	rootFolder.B_content[1].B_inodo = 0 // Apunta a sí mismo (es la raíz)
	
	// Entrada "users.txt"
	copy(rootFolder.B_content[2].B_name[:], "users.txt")
	rootFolder.B_content[2].B_inodo = 1 // Apunta al inodo de users.txt
	
	// Escribir el bloque de directorio raíz
	var rootFolderBuffer bytes.Buffer
	binary.Write(&rootFolderBuffer, binary.LittleEndian, rootFolder)
	_, err = diskFile.WriteAt(rootFolderBuffer.Bytes(), int64(blockPos))
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de directorio raíz: %v", err)
	}
	
	// Crear y escribir el contenido del archivo users.txt
	// Formato: "GID, Tipo, Grupo \nUID, Tipo, Grupo, Usuario, Contraseña \n"
	usersContent := "1, G, root \n1, U, root, root, 123 \n"
	
	// Crear el bloque de archivo para users.txt
	usersFileBlock := structs.FileBlock{}
	// Copiar el contenido al bloque (hasta 64 bytes)
	copy(usersFileBlock.B_content[:], usersContent)
	
	// Escribir el bloque de archivo
	var usersFileBuffer bytes.Buffer
	binary.Write(&usersFileBuffer, binary.LittleEndian, usersFileBlock)
	_, err = diskFile.WriteAt(usersFileBuffer.Bytes(), int64(blockPos+structs.BLOCK_SIZE))
	if err != nil {
		return fmt.Errorf("error al escribir el bloque del archivo users.txt: %v", err)
	}
	
	return nil
}

// getCurrentTimeBytes devuelve la fecha actual como un arreglo de 16 bytes
func getCurrentTimeBytes() [16]byte {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	var timeBytes [16]byte
	copy(timeBytes[:], currentTime)
	return timeBytes
}
