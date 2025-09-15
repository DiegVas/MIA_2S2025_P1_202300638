package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"strings"
)

func generateSuperblockReport(partitionID string) (string, error) {
	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el superbloque para obtener información del sistema de archivos
	superblock, err := readSuperblock(mountedPartition.Path, partitionID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Generar contenido DOT para el reporte del superbloque
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Superblock_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, fontname=\"Arial\"];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n")
	dot.WriteString("    graph [fontname=\"Arial\"];\n\n")

	// Título del reporte
	dot.WriteString(fmt.Sprintf("    title [label=\"SuperBloque Particion %s\\nen disco %s\", shape=box, style=filled, fillcolor=lightblue];\n\n",
		partitionID, mountedPartition.Path))

	// Crear la tabla del superbloque
	dot.WriteString("    superblock [label=\"")
	dot.WriteString("{SUPERBLOQUE|")

	// Información del sistema de archivos
	fsType := "Desconocido"
	switch superblock.S_filesystem_type {
	case 2:
		fsType = "EXT2"
	case 3:
		fsType = "EXT3"
	}

	dot.WriteString(fmt.Sprintf("Tipo de Sistema de Archivos|%s\\l|", fsType))
	dot.WriteString(fmt.Sprintf("Número total de inodos|%d\\l|", superblock.S_inodes_count))
	dot.WriteString(fmt.Sprintf("Número total de bloques|%d\\l|", superblock.S_blocks_count))
	dot.WriteString(fmt.Sprintf("Número de bloques libres|%d\\l|", superblock.S_free_blocks_count))
	dot.WriteString(fmt.Sprintf("Número de inodos libres|%d\\l|", superblock.S_free_inodes_count))

	// Fechas de montaje (convertir de bytes a string)
	mtimeStr := strings.TrimRight(string(superblock.S_mtime[:]), "\x00")
	if mtimeStr == "" {
		mtimeStr = "No definido"
	}
	umtimeStr := strings.TrimRight(string(superblock.S_umtime[:]), "\x00")
	if umtimeStr == "" {
		umtimeStr = "No definido"
	}

	dot.WriteString(fmt.Sprintf("Última fecha de montaje|%s\\l|", mtimeStr))

	// Número mágico
	magicValid := "Inválido"
	if superblock.S_magic == 0xEF53 {
		magicValid = "Válido (0xEF53)"
	} else {
		magicValid = fmt.Sprintf("Inválido (0x%X)", superblock.S_magic)
	}
	dot.WriteString(fmt.Sprintf("Número mágico|%s\\l|", magicValid))

	// Tamanos
	dot.WriteString(fmt.Sprintf("Tamano del inodo|%d bytes\\l|", superblock.S_inode_size))
	dot.WriteString(fmt.Sprintf("Tamano del bloque|%d bytes\\l|", superblock.S_block_size))

	// Posiciones de inicio
	dot.WriteString(fmt.Sprintf("Primer inodo libre|%d\\l|", superblock.S_first_ino))
	dot.WriteString(fmt.Sprintf("Primer bloque libre|%d\\l|", superblock.S_first_blo))
	dot.WriteString(fmt.Sprintf("Inicio bitmap de inodos|%d\\l|", superblock.S_bm_inode_start))
	dot.WriteString(fmt.Sprintf("Inicio bitmap de bloques|%d\\l|", superblock.S_bm_block_start))
	dot.WriteString(fmt.Sprintf("Inicio tabla de inodos|%d\\l|", superblock.S_inode_start))
	dot.WriteString(fmt.Sprintf("Inicio tabla de bloques|%d\\l", superblock.S_block_start))

	dot.WriteString("}\", style=filled, fillcolor=lightgray];\n\n")

	// Conectar título con la tabla
	dot.WriteString("    title -> superblock;\n\n")

	// Cerrar el grafo
	dot.WriteString("}\n")

	return dot.String(), nil
}

func readSuperblock(diskPath, partitionID string) (*structs.Superblock, error) {

	// Obtener información de la partición montada para encontrar el superbloque
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return nil, fmt.Errorf("partición no encontrada")
	}

	// Leer el MBR para encontrar la partición específica
	mbr, err := readMBRFromDisk(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Buscar la partición por nombre
	var targetPartition structs.PARTITION
	found = false

	for i, partition := range mbr.Partitions {
		fmt.Printf("Partición %d: %s\n", i, partition.Part_name)
		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		if partName == mountedPartition.Name {
			targetPartition = partition
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("partición '%s' no encontrada en el disco", mountedPartition.Name)
	}

	// Verificar que la particion tenga un tamano valido
	if targetPartition.Part_size <= 0 {
		return nil, fmt.Errorf("particion '%s' tiene un tamano invalido: %d", mountedPartition.Name, targetPartition.Part_size)
	}

	// Abrir el archivo del disco
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Verificar el tamano del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del archivo: %v", err)
	}

	// Verificar que la posición de lectura esté dentro del archivo
	readPos := int64(targetPartition.Part_start)

	if readPos+int64(structs.SUPERBLOCK_SIZE) > fileInfo.Size() {
		return nil, fmt.Errorf("posición de lectura fuera del archivo: pos=%d, size=%d, superblock_size=%d",
			readPos, fileInfo.Size(), structs.SUPERBLOCK_SIZE)
	}

	// Leer el superbloque desde la posición de inicio de la partición
	superbData := make([]byte, structs.SUPERBLOCK_SIZE)
	bytesRead, err := file.ReadAt(superbData, readPos)
	if err != nil {
		return nil, fmt.Errorf("error leyendo superbloque: %v", err)
	}

	if bytesRead != structs.SUPERBLOCK_SIZE {
		return nil, fmt.Errorf("no se pudieron leer todos los bytes del superbloque: leídos=%d, esperados=%d",
			bytesRead, structs.SUPERBLOCK_SIZE)
	}

	// Deserializar el superbloque campo por campo para evitar problemas de alineación
	var superblock structs.Superblock
	buffer := bytes.NewReader(superbData)

	// Leer campos uno por uno
	binary.Read(buffer, binary.LittleEndian, &superblock.S_filesystem_type)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_inodes_count)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_blocks_count)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_free_blocks_count)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_free_inodes_count)
	buffer.Read(superblock.S_mtime[:])
	buffer.Read(superblock.S_umtime[:])
	binary.Read(buffer, binary.LittleEndian, &superblock.S_mnt_count)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_magic)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_inode_size)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_block_size)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_first_ino)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_first_blo)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_bm_inode_start)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_bm_block_start)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_inode_start)
	binary.Read(buffer, binary.LittleEndian, &superblock.S_block_start)

	// Verificar que el superbloque tenga valores válidos
	if superblock.S_magic != structs.EXT2_MAGIC {
		return nil, fmt.Errorf("la partición '%s' no tiene un sistema de archivos EXT2 válido. Número mágico: 0x%X, esperado: 0x%X. Use 'mkfs' para formatear la partición",
			mountedPartition.Name, superblock.S_magic, structs.EXT2_MAGIC)
	}

	if superblock.S_inodes_count <= 0 {
		return nil, fmt.Errorf("número de inodos inválido: %d", superblock.S_inodes_count)
	}

	return &superblock, nil
}
