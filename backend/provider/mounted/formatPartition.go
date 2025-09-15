package mounted

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"time"
)

func FormatPartition(params *structs.MkfsParams) error {

	mountManager := GetMountManager()
	partition, found := mountManager.GetMountedPartition(params.ID)

	if !found {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s'", params.ID)
	}

	mbr, err := ReadMBR(partition.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	var targetPartition structs.PARTITION
	found = false

	for _, p := range mbr.Partitions {

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

	partitionSize := targetPartition.Part_size

	inodeCount := int32(float64(partitionSize) * 0.05 / float64(structs.INODE_SIZE))

	blockCount := (partitionSize - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	superblock := structs.Superblock{
		S_filesystem_type:   2,
		S_inodes_count:      inodeCount,
		S_blocks_count:      blockCount,
		S_free_blocks_count: blockCount - 2,
		S_free_inodes_count: inodeCount - 2,
		S_mtime:             getCurrentTimeBytes(),
		S_umtime:            [16]byte{},
		S_mnt_count:         1,
		S_magic:             structs.EXT2_MAGIC,
		S_inode_size:        structs.INODE_SIZE,
		S_block_size:        structs.BLOCK_SIZE,
		S_first_ino:         2,
		S_first_blo:         2,
	}

	// ! Calcular las posiciones de las estructuras
	startPosition := targetPartition.Part_start

	// ? Posición del superbloque
	superbPos := startPosition

	// ? Posición del bitmap de inodos
	bmInodePos := superbPos + structs.SUPERBLOCK_SIZE
	superblock.S_bm_inode_start = bmInodePos

	// ? Posición del bitmap de bloques
	bmBlockPos := bmInodePos + inodeCount
	superblock.S_bm_block_start = bmBlockPos

	// ? Posición de la tabla de inodos
	inodePos := bmBlockPos + blockCount
	superblock.S_inode_start = inodePos

	// ? Posición de la tabla de bloques
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)
	superblock.S_block_start = blockPos

	diskFile, err := os.OpenFile(partition.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	var superbBuffer bytes.Buffer
	binary.Write(&superbBuffer, binary.LittleEndian, superblock)
	_, err = diskFile.WriteAt(superbBuffer.Bytes(), int64(superbPos))
	if err != nil {
		return fmt.Errorf("error al escribir el superbloque: %v", err)
	}

	inodeBitmap := make([]byte, inodeCount)
	for i := range inodeBitmap {
		inodeBitmap[i] = 0
	}

	inodeBitmap[0] = 1
	inodeBitmap[1] = 1

	_, err = diskFile.WriteAt(inodeBitmap, int64(bmInodePos))
	if err != nil {
		return fmt.Errorf("error al escribir el bitmap de inodos: %v", err)
	}

	blockBitmap := make([]byte, blockCount)
	for i := range blockBitmap {
		blockBitmap[i] = 0
	}

	blockBitmap[0] = 1
	blockBitmap[1] = 1

	// ! Escribir el bitmap de bloques
	_, err = diskFile.WriteAt(blockBitmap, int64(bmBlockPos))
	if err != nil {
		return fmt.Errorf("error al escribir el bitmap de bloques: %v", err)
	}

	rootInode := structs.Inode{
		I_uid:   0,
		I_gid:   0,
		I_size:  structs.FOLDER_SIZE,
		I_atime: getCurrentTimeBytes(),
		I_ctime: getCurrentTimeBytes(),
		I_mtime: getCurrentTimeBytes(),
		I_block: [15]int32{0}, // Caso especial: directorio raíz usa bloque 0
		I_type:  1,            // 1 = directorio
		I_perm:  664,
	}

	var rootInodeBuffer bytes.Buffer
	binary.Write(&rootInodeBuffer, binary.LittleEndian, rootInode)
	_, err = diskFile.WriteAt(rootInodeBuffer.Bytes(), int64(inodePos))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo raíz: %v", err)
	}

	usersInode := structs.Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len("1, G, root \n1, U, root, root, 123 \n")),
		I_atime: getCurrentTimeBytes(),
		I_ctime: getCurrentTimeBytes(),
		I_mtime: getCurrentTimeBytes(),
		I_block: [15]int32{1}, // El archivo users.txt está en el bloque lógico 1
		I_type:  0,            // 0 = archivo
		I_perm:  664,
	}

	var usersInodeBuffer bytes.Buffer
	binary.Write(&usersInodeBuffer, binary.LittleEndian, usersInode)
	_, err = diskFile.WriteAt(usersInodeBuffer.Bytes(), int64(inodePos+structs.INODE_SIZE))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo de users.txt: %v", err)
	}

	// ! Crear el bloque para el directorio raíz
	rootFolder := structs.FolderBlock{}

	copy(rootFolder.B_content[0].B_name[:], ".")
	rootFolder.B_content[0].B_inodo = 0

	copy(rootFolder.B_content[1].B_name[:], "..")
	rootFolder.B_content[1].B_inodo = 0

	copy(rootFolder.B_content[2].B_name[:], "users.txt")
	rootFolder.B_content[2].B_inodo = 1

	var rootFolderBuffer bytes.Buffer
	binary.Write(&rootFolderBuffer, binary.LittleEndian, rootFolder)
	_, err = diskFile.WriteAt(rootFolderBuffer.Bytes(), int64(blockPos)) // Bloque 0
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de directorio raíz: %v", err)
	}

	usersContent := "1, G, root \n1, U, root, root, 123 \n"

	usersFileBlock := structs.FileBlock{}

	copy(usersFileBlock.B_content[:], usersContent)

	var usersFileBuffer bytes.Buffer
	binary.Write(&usersFileBuffer, binary.LittleEndian, usersFileBlock)
	_, err = diskFile.WriteAt(usersFileBuffer.Bytes(), int64(blockPos+structs.BLOCK_SIZE)) // Bloque 1
	if err != nil {
		return fmt.Errorf("error al escribir el bloque del archivo users.txt: %v", err)
	}

	return nil
}

func getCurrentTimeBytes() [16]byte {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	var timeBytes [16]byte
	copy(timeBytes[:], currentTime)
	return timeBytes
}
