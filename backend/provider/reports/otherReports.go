package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/provider/files"
	"main/provider/mounted"
	"main/structs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// generateInodeReport genera el reporte de inodos utilizados
func generateInodeReport(partitionID string) (string, error) {

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

	// Leer bitmap de inodos para identificar cuáles están en uso
	inodeBitmap, err := readInodeBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	// Leer todos los inodos utilizados
	usedInodes, err := readUsedInodes(mountedPartition.Path, superblock, inodeBitmap)
	if err != nil {
		return "", fmt.Errorf("error leyendo inodos: %v", err)
	}

	if len(usedInodes) == 0 {
		return "", fmt.Errorf("no se encontraron inodos utilizados")
	}

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Inode_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	titleLabel := fmt.Sprintf("INODE Report\\nPartición: %s", escapeLabel(partitionID))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Generar nodos para cada inodo utilizado
	for i, inodeInfo := range usedInodes {
		inodeLabel := generateInodeLabel(i, inodeInfo)

		// Determinar color según el tipo de inodo
		var fillColor string
		if inodeInfo.Inode.I_type == 1 {
			fillColor = "lightgreen" // Directorios en verde claro
		} else {
			fillColor = "lightcyan" // Archivos en cian claro
		}

		dot.WriteString(fmt.Sprintf("    inode_%d [label=\"%s\", fillcolor=%s];\n", i, inodeLabel, fillColor))
	}

	// Generar conexiones para crear la lista enlazada
	dot.WriteString("\n    // Conexiones de la lista enlazada\n")
	dot.WriteString("    title -> inode_0;\n")

	for i := 0; i < len(usedInodes)-1; i++ {
		dot.WriteString(fmt.Sprintf("    inode_%d -> inode_%d;\n", i, i+1))
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

// readSuperblock lee el superbloque de la partición
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

	// Verificar que la partición tenga un tamaño válido
	if targetPartition.Part_size <= 0 {
		return nil, fmt.Errorf("partición '%s' tiene un tamaño inválido: %d", mountedPartition.Name, targetPartition.Part_size)
	}

	// Abrir el archivo del disco
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Verificar el tamaño del archivo
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

// readInodeBitmap lee el bitmap de inodos
func readInodeBitmap(diskPath string, superblock *structs.Superblock) ([]byte, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Leer el bitmap de inodos
	bitmapSize := superblock.S_inodes_count
	bitmap := make([]byte, bitmapSize)

	_, err = file.ReadAt(bitmap, int64(superblock.S_bm_inode_start))
	if err != nil {
		return nil, fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	return bitmap, nil
}

// readUsedInodes lee todos los inodos que están marcados como utilizados
func readUsedInodes(diskPath string, superblock *structs.Superblock, bitmap []byte) ([]inodeInfo, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	var usedInodes []inodeInfo

	// Recorrer el bitmap para encontrar inodos utilizados
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		if i < int32(len(bitmap)) && bitmap[i] == 1 {
			// Este inodo está en uso, leerlo
			inodeData := make([]byte, structs.INODE_SIZE)
			inodePos := int64(superblock.S_inode_start) + int64(i)*int64(structs.INODE_SIZE)

			_, err = file.ReadAt(inodeData, inodePos)
			if err != nil {
				continue // Saltar este inodo si hay error
			}

			// Deserializar el inodo
			var inode structs.Inode
			buffer := bytes.NewReader(inodeData)
			err = binary.Read(buffer, binary.LittleEndian, &inode)
			if err != nil {
				continue // Saltar este inodo si hay error
			}

			// Agregar información del inodo
			usedInodes = append(usedInodes, inodeInfo{
				Number: i,
				Inode:  inode,
			})
		}
	}

	return usedInodes, nil
}

// generateInodeLabel genera la etiqueta para mostrar la información del inodo
func generateInodeLabel(index int, inodeInfo inodeInfo) string {

	inode := inodeInfo.Inode

	// Formatear las fechas
	atime := formatInodeTime(inode.I_atime)
	ctime := formatInodeTime(inode.I_ctime)
	mtime := formatInodeTime(inode.I_mtime)

	// Determinar el tipo de inodo
	inodeType := "Archivo"
	if inode.I_type == 1 {
		inodeType = "Directorio"
	}

	// Analizar bloques con más detalle
	var directBlocks []string
	var indirectBlocks []string
	var indirectSimple, indirectDoble, indirectTriple int32

	// Bloques directos (0-11)
	for i := 0; i < 12 && i < len(inode.I_block); i++ {
		if inode.I_block[i] != 0 {
			directBlocks = append(directBlocks, fmt.Sprintf("%d", inode.I_block[i]))
		}
	}

	// Bloques indirectos
	if len(inode.I_block) > 12 && inode.I_block[12] != 0 {
		indirectSimple = inode.I_block[12]
		indirectBlocks = append(indirectBlocks, fmt.Sprintf("IS:%d", indirectSimple))
	}
	if len(inode.I_block) > 13 && inode.I_block[13] != 0 {
		indirectDoble = inode.I_block[13]
		indirectBlocks = append(indirectBlocks, fmt.Sprintf("ID:%d", indirectDoble))
	}
	if len(inode.I_block) > 14 && inode.I_block[14] != 0 {
		indirectTriple = inode.I_block[14]
		indirectBlocks = append(indirectBlocks, fmt.Sprintf("IT:%d", indirectTriple))
	}

	// Formatear información de bloques
	var blockInfo string
	if len(directBlocks) > 0 {
		if len(directBlocks) > 3 {
			blockInfo = fmt.Sprintf("Directos: [%s...%d más]",
				strings.Join(directBlocks[:3], ","), len(directBlocks)-3)
		} else {
			blockInfo = fmt.Sprintf("Directos: [%s]", strings.Join(directBlocks, ","))
		}
	}

	if len(indirectBlocks) > 0 {
		if blockInfo != "" {
			blockInfo += "\\l"
		}
		blockInfo += fmt.Sprintf("Indirectos: [%s]", strings.Join(indirectBlocks, ","))
	}

	if blockInfo == "" {
		blockInfo = "Sin bloques asignados"
	}

	// Formatear permisos en octal
	permStr := fmt.Sprintf("%04o", inode.I_perm)

	// Generar la etiqueta con formato mejorado
	label := fmt.Sprintf("INODO %d\\n(%s)|"+
		"{UID: %d | GID: %d}|"+
		"{Tamaño: %d bytes | Permisos: %s}|"+
		"{%s}|"+
		"{A-time: %s}|"+
		"{C-time: %s}|"+
		"{M-time: %s}",
		inodeInfo.Number, inodeType,
		inode.I_uid, inode.I_gid,
		inode.I_size, permStr,
		blockInfo,
		escapeLabel(atime),
		escapeLabel(ctime),
		escapeLabel(mtime))

	return label
}

// formatInodeTime formatea el tiempo del inodo
func formatInodeTime(timeBytes [16]byte) string {
	timeStr := strings.TrimRight(string(timeBytes[:]), "\x00")
	if timeStr == "" {
		return "N/A"
	}

	// Intentar parsear diferentes formatos de fecha
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}

	return timeStr
}

// inodeInfo estructura auxiliar para almacenar información del inodo
type inodeInfo struct {
	Number int32
	Inode  structs.Inode
}

// generateBlockReport genera el reporte de bloques
func generateBlockReport(partitionID string) (string, error) {

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

	// Leer bitmap de bloques para identificar cuáles están en uso
	blockBitmap, err := readBlockBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	// Leer todos los bloques utilizados
	usedBlocks, err := readUsedBlocks(mountedPartition.Path, superblock, blockBitmap)
	if err != nil {
		return "", fmt.Errorf("error leyendo bloques: %v", err)
	}

	if len(usedBlocks) == 0 {
		return "", fmt.Errorf("no se encontraron bloques utilizados")
	}

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Block_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record, style=filled];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    node [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	titleLabel := fmt.Sprintf("BLOCK Report\\nPartición: %s", escapeLabel(partitionID))
	dot.WriteString(fmt.Sprintf("    title [label=\"%s\", fillcolor=lightblue, fontsize=14];\n\n", titleLabel))

	// Generar nodos para cada bloque utilizado
	for i, block := range usedBlocks {
		blockLabel := generateBlockLabel(i, block)
		dot.WriteString(fmt.Sprintf("    block_%d [label=\"%s\", fillcolor=lightyellow];\n", i, blockLabel))
	}

	// Generar conexiones para crear la lista enlazada
	dot.WriteString("\n    // Conexiones de la lista enlazada\n")
	dot.WriteString("    title -> block_0;\n")

	for i := 0; i < len(usedBlocks)-1; i++ {
		dot.WriteString(fmt.Sprintf("    block_%d -> block_%d;\n", i, i+1))
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

// generateBMInodeTextReport genera el reporte del bitmap de inodos como archivo de texto
func generateBMInodeTextReport(params *structs.RepParams) (string, error) {
	// Obtener la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(params.ID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", params.ID)
	}

	// Leer el superbloque
	superblock, err := readSuperblock(mountedPartition.Path, params.ID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Leer bitmap de inodos
	inodeBitmap, err := readInodeBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar contenido del archivo de texto
	var content strings.Builder
	content.WriteString("REPORTE DE BITMAP DE INODOS\n")
	content.WriteString("===========================\n")
	content.WriteString(fmt.Sprintf("Partición: %s\n", params.ID))
	content.WriteString(fmt.Sprintf("Total de inodos: %d\n", superblock.S_inodes_count))
	content.WriteString(fmt.Sprintf("Inodos libres: %d\n", superblock.S_free_inodes_count))
	content.WriteString(fmt.Sprintf("Inodos utilizados: %d\n\n", superblock.S_inodes_count-superblock.S_free_inodes_count))
	content.WriteString("BITMAP (0=libre, 1=ocupado):\n")
	content.WriteString("============================\n")

	// Mostrar bitmap con 20 registros por línea
	for i := 0; i < len(inodeBitmap); i += 20 {
		// Línea de números de inodo
		content.WriteString(fmt.Sprintf("%4d: ", i))

		end := i + 20
		if end > len(inodeBitmap) {
			end = len(inodeBitmap)
		}

		// Mostrar los bits
		for j := i; j < end; j++ {
			content.WriteString(fmt.Sprintf("%d", inodeBitmap[j]))
		}
		content.WriteString("\n")
	}

	// Escribir archivo
	err = os.WriteFile(params.Path, []byte(content.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return params.Path, nil
}

// generateBMBlockTextReport genera el reporte del bitmap de bloques como archivo de texto
func generateBMBlockTextReport(params *structs.RepParams) (string, error) {
	// Obtener la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(params.ID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", params.ID)
	}

	// Leer el superbloque
	superblock, err := readSuperblock(mountedPartition.Path, params.ID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Leer bitmap de bloques
	blockBitmap, err := readBlockBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar contenido del archivo de texto
	var content strings.Builder
	content.WriteString("REPORTE DE BITMAP DE BLOQUES\n")
	content.WriteString("============================\n")
	content.WriteString(fmt.Sprintf("Partición: %s\n", params.ID))
	content.WriteString(fmt.Sprintf("Total de bloques: %d\n", superblock.S_blocks_count))
	content.WriteString(fmt.Sprintf("Bloques libres: %d\n", superblock.S_free_blocks_count))
	content.WriteString(fmt.Sprintf("Bloques utilizados: %d\n\n", superblock.S_blocks_count-superblock.S_free_blocks_count))
	content.WriteString("BITMAP (0=libre, 1=ocupado):\n")
	content.WriteString("=============================\n")

	// Mostrar bitmap con 20 registros por línea
	for i := 0; i < len(blockBitmap); i += 20 {
		// Línea de números de bloque
		content.WriteString(fmt.Sprintf("%4d: ", i))

		end := i + 20
		if end > len(blockBitmap) {
			end = len(blockBitmap)
		}

		// Mostrar los bits
		for j := i; j < end; j++ {
			content.WriteString(fmt.Sprintf("%d", blockBitmap[j]))
		}
		content.WriteString("\n")
	}

	// Escribir archivo
	err = os.WriteFile(params.Path, []byte(content.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return params.Path, nil
}

// generateBMBlockReport genera el reporte del bitmap de bloques
func generateBMBlockReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph BM_Block_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"BM Block Report\" [label=\"Bitmap de Bloques\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateTreeReport genera el reporte del árbol completo del sistema de archivos ext2
func generateTreeReport(partitionID string) (string, error) {
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

	// Abrir el archivo del disco
	file, err := os.Open(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Generar contenido DOT
	var dot strings.Builder

	// Iniciar el grafo
	dot.WriteString("digraph Tree_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=plaintext];\n")
	dot.WriteString("    graph [fontname=\"Arial\", fontsize=12];\n")
	dot.WriteString("    edge [fontname=\"Arial\"];\n\n")

	// Título del reporte
	dot.WriteString(fmt.Sprintf("    title [label=\"Reporte Árbol del Sistema de Archivos\\nPartición: %s\", shape=box, style=filled, fillcolor=lightblue];\n\n", partitionID))

	// Generar el árbol completo del sistema de archivos
	err = generateFileSystemTree(file, superblock, &dot)
	if err != nil {
		return "", fmt.Errorf("error generando árbol del sistema de archivos: %v", err)
	}

	dot.WriteString("}\n")

	return dot.String(), nil
}

// generateFileSystemTree genera el árbol completo del sistema de archivos
func generateFileSystemTree(file *os.File, superblock *structs.Superblock, dot *strings.Builder) error {
	// Leer bitmap de inodos para identificar cuáles están en uso
	inodeBitmap, err := readInodeBitmapFromFile(file, superblock)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	// Leer bitmap de bloques
	blockBitmap, err := readBlockBitmapFromFile(file, superblock)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	// Contador para nodos únicos
	nodeCounter := 0

	// Generar inodos utilizados
	err = generateUsedInodes(file, superblock, inodeBitmap, dot, &nodeCounter)
	if err != nil {
		return fmt.Errorf("error generando inodos: %v", err)
	}

	// Generar bloques utilizados
	err = generateUsedBlocks(file, superblock, blockBitmap, dot, &nodeCounter)
	if err != nil {
		return fmt.Errorf("error generando bloques: %v", err)
	}

	// Generar conexiones entre inodos y bloques
	err = generateInodeBlockConnections(file, superblock, inodeBitmap, blockBitmap, dot)
	if err != nil {
		return fmt.Errorf("error generando conexiones: %v", err)
	}

	return nil
}

// readInodeBitmapFromFile lee el bitmap de inodos desde el archivo
func readInodeBitmapFromFile(file *os.File, superblock *structs.Superblock) ([]byte, error) {
	bitmapSize := superblock.S_inodes_count
	bitmap := make([]byte, bitmapSize)

	_, err := file.ReadAt(bitmap, int64(superblock.S_bm_inode_start))
	if err != nil {
		return nil, fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	return bitmap, nil
}

// readBlockBitmapFromFile lee el bitmap de bloques desde el archivo
func readBlockBitmapFromFile(file *os.File, superblock *structs.Superblock) ([]byte, error) {
	bitmapSize := superblock.S_blocks_count
	bitmap := make([]byte, bitmapSize)

	_, err := file.ReadAt(bitmap, int64(superblock.S_bm_block_start))
	if err != nil {
		return nil, fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	return bitmap, nil
}

// generateUsedInodes genera los nodos de inodos utilizados
func generateUsedInodes(file *os.File, superblock *structs.Superblock, bitmap []byte, dot *strings.Builder, nodeCounter *int) error {
	// Recorrer el bitmap para encontrar inodos utilizados
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		if i < int32(len(bitmap)) && bitmap[i] == 1 {
			// Este inodo está en uso, leerlo
			inode, err := readInodeAt(file, i, superblock.S_inode_start)
			if err != nil {
				continue // Saltar este inodo si hay error
			}

			// Generar tabla HTML para el inodo
			inodeTable := generateInodeTableHTML(i, inode)

			dot.WriteString(fmt.Sprintf("    inode_%d [label=<%s>];\n", i, inodeTable))
			(*nodeCounter)++
		}
	}

	return nil
}

// generateUsedBlocks genera los nodos de bloques utilizados
func generateUsedBlocks(file *os.File, superblock *structs.Superblock, bitmap []byte, dot *strings.Builder, nodeCounter *int) error {
	// Recorrer el bitmap para encontrar bloques utilizados
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		if i < int32(len(bitmap)) && bitmap[i] == 1 {
			// Este bloque está en uso, leerlo
			blockData, blockType, err := readBlockAt(file, i, superblock)
			if err != nil {
				continue // Saltar este bloque si hay error
			}

			// Generar tabla HTML para el bloque
			blockTable := generateBlockTableHTML(i, blockData, blockType)

			dot.WriteString(fmt.Sprintf("    block_%d [label=<%s>];\n", i, blockTable))
			(*nodeCounter)++
		}
	}

	return nil
}

// generateInodeBlockConnections genera las conexiones entre inodos y bloques
func generateInodeBlockConnections(file *os.File, superblock *structs.Superblock, inodeBitmap, blockBitmap []byte, dot *strings.Builder) error {
	// Recorrer inodos utilizados y crear conexiones con sus bloques
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		if i < int32(len(inodeBitmap)) && inodeBitmap[i] == 1 {
			// Este inodo está en uso, leer sus bloques
			inode, err := readInodeAt(file, i, superblock.S_inode_start)
			if err != nil {
				continue
			}

			// Conectar con bloques directos
			for j := 0; j < 12; j++ {
				if inode.I_block[j] != 0 && inode.I_block[j] < int32(len(blockBitmap)) && blockBitmap[inode.I_block[j]] == 1 {
					dot.WriteString(fmt.Sprintf("    inode_%d:block_%d -> block_%d;\n", i, j, inode.I_block[j]))
				}
			}

			// Conectar con bloque de punteros indirectos (bloque 12)
			if inode.I_block[12] != 0 && inode.I_block[12] < int32(len(blockBitmap)) && blockBitmap[inode.I_block[12]] == 1 {
				dot.WriteString(fmt.Sprintf("    inode_%d:block_12 -> block_%d;\n", i, inode.I_block[12]))

				// Leer bloque de punteros y conectar con bloques apuntados
				err = generateIndirectBlockConnections(file, superblock, inode.I_block[12], blockBitmap, dot)
				if err == nil {
					// Las conexiones indirectas se manejan dentro de generateIndirectBlockConnections
				}
			}
		}
	}

	return nil
}

// generateIndirectBlockConnections genera conexiones para bloques de punteros indirectos
func generateIndirectBlockConnections(file *os.File, superblock *structs.Superblock, blockNum int32, blockBitmap []byte, dot *strings.Builder) error {
	// Leer el bloque de punteros
	pointerBlock, err := readPointerBlockAt(file, blockNum, superblock)
	if err != nil {
		return err
	}

	// Conectar con cada bloque apuntado
	for i, pointer := range pointerBlock.B_pointers {
		if pointer != 0 && pointer < int32(len(blockBitmap)) && blockBitmap[pointer] == 1 {
			dot.WriteString(fmt.Sprintf("    block_%d:ptr_%d -> block_%d;\n", blockNum, i, pointer))
		}
	}

	return nil
}

// generateInodeTableHTML genera tabla HTML para mostrar información del inodo
func generateInodeTableHTML(inodeNum int32, inode *structs.Inode) string {
	inodeType := "Archivo"
	var bgColor string
	if inode.I_type == 1 {
		inodeType = "Directorio"
		bgColor = "lightgreen"
	} else {
		bgColor = "lightcyan"
	}

	atime := formatInodeTime(inode.I_atime)
	ctime := formatInodeTime(inode.I_ctime)
	mtime := formatInodeTime(inode.I_mtime)

	var table strings.Builder
	table.WriteString("<table border=\"1\" cellborder=\"1\" cellspacing=\"0\">")
	table.WriteString(fmt.Sprintf("<tr><td colspan=\"2\" bgcolor=\"%s\"><b>INODO %d</b></td></tr>", bgColor, inodeNum))
	table.WriteString(fmt.Sprintf("<tr><td>Tipo</td><td>%s</td></tr>", inodeType))
	table.WriteString(fmt.Sprintf("<tr><td>UID</td><td>%d</td></tr>", inode.I_uid))
	table.WriteString(fmt.Sprintf("<tr><td>GID</td><td>%d</td></tr>", inode.I_gid))
	table.WriteString(fmt.Sprintf("<tr><td>Tamaño</td><td>%d bytes</td></tr>", inode.I_size))
	table.WriteString(fmt.Sprintf("<tr><td>Permisos</td><td>%o</td></tr>", inode.I_perm))
	table.WriteString(fmt.Sprintf("<tr><td>A-time</td><td>%s</td></tr>", atime))
	table.WriteString(fmt.Sprintf("<tr><td>C-time</td><td>%s</td></tr>", ctime))
	table.WriteString(fmt.Sprintf("<tr><td>M-time</td><td>%s</td></tr>", mtime))

	// Mostrar bloques directos e indirectos
	for i := 0; i < 15; i++ {
		if inode.I_block[i] != 0 {
			blockType := "Directo"
			if i == 12 {
				blockType = "Indirecto Simple"
			} else if i == 13 {
				blockType = "Indirecto Doble"
			} else if i == 14 {
				blockType = "Indirecto Triple"
			}
			table.WriteString(fmt.Sprintf("<tr><td port=\"block_%d\">Bloque %d (%s)</td><td>%d</td></tr>", i, i, blockType, inode.I_block[i]))
		}
	}

	table.WriteString("</table>")
	return table.String()
}

// generateBlockTableHTML genera tabla HTML para mostrar información del bloque
func generateBlockTableHTML(blockNum int32, blockData []byte, blockType string) string {
	var table strings.Builder

	// Determinar color según el tipo de bloque
	var bgColor string
	switch blockType {
	case "Directorio":
		bgColor = "lightgreen"
	case "Archivo":
		bgColor = "lightcyan"
	case "Punteros":
		bgColor = "lightyellow"
	default:
		bgColor = "lightgray"
	}

	table.WriteString("<table border=\"1\" cellborder=\"1\" cellspacing=\"0\">")
	table.WriteString(fmt.Sprintf("<tr><td colspan=\"2\" bgcolor=\"%s\"><b>BLOQUE %d</b></td></tr>", bgColor, blockNum))
	table.WriteString(fmt.Sprintf("<tr><td>Tipo</td><td>%s</td></tr>", blockType))

	switch blockType {
	case "Directorio":
		err := generateDirectoryBlockTable(&table, blockData)
		if err != nil {
			table.WriteString(fmt.Sprintf("<tr><td>Error</td><td>%s</td></tr>", err.Error()))
		}
	case "Archivo":
		generateFileBlockTable(&table, blockData)
	case "Punteros":
		err := generatePointerBlockTable(&table, blockData)
		if err != nil {
			table.WriteString(fmt.Sprintf("<tr><td>Error</td><td>%s</td></tr>", err.Error()))
		}
	default:
		table.WriteString("<tr><td>Contenido</td><td>Datos binarios</td></tr>")
	}

	table.WriteString("</table>")
	return table.String()
}

// generateDirectoryBlockTable genera la tabla para un bloque de directorio
func generateDirectoryBlockTable(table *strings.Builder, blockData []byte) error {
	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &folderBlock)
	if err != nil {
		return err
	}

	for i, content := range folderBlock.B_content {
		entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if entryName != "" {
			table.WriteString(fmt.Sprintf("<tr><td port=\"entry_%d\">Entrada %d</td><td>%s (Inodo: %d)</td></tr>", i, i, entryName, content.B_inodo))
		}
	}

	return nil
}

// generateFileBlockTable genera la tabla para un bloque de archivo
func generateFileBlockTable(table *strings.Builder, blockData []byte) {
	var fileBlock structs.FileBlock
	buffer := bytes.NewReader(blockData)
	binary.Read(buffer, binary.LittleEndian, &fileBlock)

	content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	table.WriteString(fmt.Sprintf("<tr><td>Contenido</td><td>%s</td></tr>", content))
}

// generatePointerBlockTable genera la tabla para un bloque de punteros
func generatePointerBlockTable(table *strings.Builder, blockData []byte) error {
	var pointerBlock structs.PointerBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err != nil {
		return err
	}

	ptrCount := 0
	for i, pointer := range pointerBlock.B_pointers {
		if pointer != 0 && ptrCount < 5 { // Mostrar solo los primeros 5 punteros
			table.WriteString(fmt.Sprintf("<tr><td port=\"ptr_%d\">Puntero %d</td><td>%d</td></tr>", i, i, pointer))
			ptrCount++
		}
	}

	// Contar total de punteros válidos
	totalValid := 0
	for _, pointer := range pointerBlock.B_pointers {
		if pointer != 0 {
			totalValid++
		}
	}

	if totalValid > 5 {
		table.WriteString(fmt.Sprintf("<tr><td colspan=\"2\">... y %d punteros más</td></tr>", totalValid-5))
	}

	return nil
}

// readBlockAt lee un bloque específico y determina su tipo
func readBlockAt(file *os.File, blockNum int32, superblock *structs.Superblock) ([]byte, string, error) {
	blockData := make([]byte, structs.BLOCK_SIZE)

	var blockPos int32
	if superblock.S_block_start == 0 {
		blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
		blockPos += (blockNum * structs.BLOCK_SIZE)
	} else {
		blockPos = superblock.S_block_start + (blockNum * structs.BLOCK_SIZE)
	}

	_, err := file.ReadAt(blockData, int64(blockPos))
	if err != nil {
		return nil, "", err
	}

	// Determinar tipo de bloque basado en su contenido
	blockType := determineBlockTypeForTree(blockData)

	return blockData, blockType, nil
}

// readPointerBlockAt lee un bloque de punteros específico
func readPointerBlockAt(file *os.File, blockNum int32, superblock *structs.Superblock) (*structs.PointerBlock, error) {
	blockData := make([]byte, structs.BLOCK_SIZE)

	var blockPos int32
	if superblock.S_block_start == 0 {
		blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
		blockPos += (blockNum * structs.BLOCK_SIZE)
	} else {
		blockPos = superblock.S_block_start + (blockNum * structs.BLOCK_SIZE)
	}

	_, err := file.ReadAt(blockData, int64(blockPos))
	if err != nil {
		return nil, err
	}

	var pointerBlock structs.PointerBlock
	buffer := bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err != nil {
		return nil, err
	}

	return &pointerBlock, nil
}

// determineBlockTypeForTree determina el tipo de bloque basado en su contenido para el reporte tree
func determineBlockTypeForTree(blockData []byte) string {
	// Usar la misma lógica mejorada que para el reporte de bloques
	blockType, _ := determineBlockType(blockData)

	// Convertir los tipos internos a los nombres usados en el reporte tree
	switch blockType {
	case "FOLDER":
		return "Directorio"
	case "POINTER":
		return "Punteros"
	case "FILE":
		return "Archivo"
	default:
		return "Archivo"
	}
}

// generateSuperblockReport genera el reporte del superbloque
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
	dot.WriteString(fmt.Sprintf("    title [label=\"SuperBloque Partición %s\\nen disco %s\", shape=box, style=filled, fillcolor=lightblue];\n\n",
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
	dot.WriteString(fmt.Sprintf("Última fecha de desmontaje|%s\\l|", umtimeStr))
	dot.WriteString(fmt.Sprintf("Número de montajes|%d\\l|", superblock.S_mnt_count))

	// Número mágico
	magicValid := "Inválido"
	if superblock.S_magic == 0xEF53 {
		magicValid = "Válido (0xEF53)"
	} else {
		magicValid = fmt.Sprintf("Inválido (0x%X)", superblock.S_magic)
	}
	dot.WriteString(fmt.Sprintf("Número mágico|%s\\l|", magicValid))

	// Tamaños
	dot.WriteString(fmt.Sprintf("Tamaño del inodo|%d bytes\\l|", superblock.S_inode_size))
	dot.WriteString(fmt.Sprintf("Tamaño del bloque|%d bytes\\l|", superblock.S_block_size))

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

// generateFileReport genera el reporte de un archivo específico
func generateFileReport(partitionID, filePath string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph File_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString(fmt.Sprintf("    \"File Report\" [label=\"Archivo: %s\\nPartición: %s\"];\n",
		filePath, partitionID))
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateFileTextReport genera el reporte del contenido de un archivo como texto
func generateFileTextReport(params *structs.RepParams) (string, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Usar la función ReadFileContent existente para leer el contenido del archivo
	// Nota: Esta función requiere que haya una sesión activa y usa la partición montada de la sesión
	fileContent, err := files.ReadFileContent(params.PathFileLS)
	if err != nil {
		return "", fmt.Errorf("error leyendo el archivo '%s': %v", params.PathFileLS, err)
	}

	// Generar contenido del reporte
	var content strings.Builder
	content.WriteString("REPORTE DE ARCHIVO\n")
	content.WriteString("==================\n")
	content.WriteString(fmt.Sprintf("Partición: %s\n", params.ID))
	content.WriteString(fmt.Sprintf("Archivo: %s\n", params.PathFileLS))
	content.WriteString(fmt.Sprintf("Tamaño: %d bytes\n", len(fileContent)))

	// Agregar fecha actual como fecha del reporte
	currentTime := time.Now()
	content.WriteString(fmt.Sprintf("Fecha del reporte: %s\n", currentTime.Format("2006-01-02 15:04:05")))

	content.WriteString("\nCONTENIDO DEL ARCHIVO:\n")
	content.WriteString("=====================\n")
	content.WriteString(fileContent)

	// Si el archivo no termina en nueva línea, agregarla
	if !strings.HasSuffix(fileContent, "\n") {
		content.WriteString("\n")
	}

	// Escribir archivo
	err = os.WriteFile(params.Path, []byte(content.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo de reporte: %v", err)
	}

	return params.Path, nil
}

// generateLSReport genera el reporte de listado de directorio
func generateLSReport(partitionID, dirPath string) (string, error) {
	// Validar parámetros
	if dirPath == "" {
		dirPath = "/"
	}

	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el superbloque
	superblock, err := readSuperblock(mountedPartition.Path, partitionID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Abrir el archivo del disco
	file, err := os.Open(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Buscar el inodo del directorio especificado
	dirInodeNum, err := findInodeByPath(file, dirPath, superblock)
	if err != nil {
		return "", fmt.Errorf("error encontrando directorio '%s': %v", dirPath, err)
	}

	// Leer el inodo del directorio
	dirInode, err := readInodeAt(file, dirInodeNum, superblock.S_inode_start)
	if err != nil {
		return "", fmt.Errorf("error leyendo inodo del directorio: %v", err)
	}

	// Verificar que sea un directorio
	if dirInode.I_type != 1 {
		return "", fmt.Errorf("'%s' no es un directorio", dirPath)
	}

	// Obtener entradas del directorio
	entries, err := getDirectoryEntries(file, dirInode, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo entradas del directorio: %v", err)
	}

	// Debug: Imprimir información del directorio
	for i, entry := range entries {
		fmt.Printf("DEBUG LS: Entrada %d: %s (inodo %d)\n", i, entry.Name, entry.InodeNum)
	}

	// Generar contenido DOT
	var dot strings.Builder
	dot.WriteString("digraph LS_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=plaintext];\n")
	dot.WriteString("    edge [style=invisible];\n")

	// Leer archivo users.txt para obtener información de usuarios
	usersInfo, err := readUsersFile(file, superblock)
	if err != nil {
		// Si no existe users.txt, usar UIDs/GIDs numéricos
		usersInfo = make(map[int32]UserInfo) // Crear mapa vacío si no se puede leer
	}

	// Título del reporte
	dot.WriteString("    title [label=<<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"lightblue\">\n")
	dot.WriteString(fmt.Sprintf("        <TR><TD COLSPAN=\"7\"><B>Reporte LS - Directorio: %s</B></TD></TR>\n", dirPath))
	dot.WriteString("        <TR>\n")
	dot.WriteString("            <TD><B>Permisos</B></TD>\n")
	dot.WriteString("            <TD><B>Owner</B></TD>\n")
	dot.WriteString("            <TD><B>Group</B></TD>\n")
	dot.WriteString("            <TD><B>Size</B></TD>\n")
	dot.WriteString("            <TD><B>Date</B></TD>\n")
	dot.WriteString("            <TD><B>Type</B></TD>\n")
	dot.WriteString("            <TD><B>Name</B></TD>\n")
	dot.WriteString("        </TR>\n")

	// Procesar cada entrada
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}

		// Leer el inodo de la entrada
		entryInode, err := readInodeAt(file, entry.InodeNum, superblock.S_inode_start)
		if err != nil {
			continue // Saltar entradas con error
		}

		// Obtener información de la entrada
		perms := formatPermissions(entryInode.I_perm)

		// Obtener nombres de usuario y grupo desde users.txt
		ownerName := fmt.Sprintf("%d", entryInode.I_uid)
		groupName := fmt.Sprintf("%d", entryInode.I_gid)
		if userInfo, exists := usersInfo[entryInode.I_uid]; exists {
			ownerName = userInfo.Username
			groupName = userInfo.Group
		}

		size := fmt.Sprintf("%d", entryInode.I_size)

		// Formatear fecha de creación
		creationTime := strings.TrimRight(string(entryInode.I_ctime[:]), "\x00")
		date := "N/A"

		if creationTime != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", creationTime); err == nil {
				date = t.Format("2006-01-02")
			}
		}

		// Determinar tipo
		fileType := "file"
		if entryInode.I_type == 1 {
			fileType = "dir"
		}

		// Agregar fila a la tabla (sin columna de Time)
		dot.WriteString("        <TR>\n")
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", perms))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", ownerName))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", groupName))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", size))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", date))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", fileType))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", entry.Name))
		dot.WriteString("        </TR>\n")
	}

	dot.WriteString("    </TABLE>>];\n")
	dot.WriteString("}\n")

	return dot.String(), nil
}

// readBlockBitmap lee el bitmap de bloques
func readBlockBitmap(diskPath string, superblock *structs.Superblock) ([]byte, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Calcular la posición del bitmap de bloques
	bmBlockPos := superblock.S_bm_inode_start + superblock.S_inodes_count

	bitmap := make([]byte, superblock.S_blocks_count)
	_, err = file.ReadAt(bitmap, int64(bmBlockPos))
	if err != nil {
		return nil, err
	}

	return bitmap, nil
}

// blockInfo estructura auxiliar para almacenar información del bloque
type blockInfo struct {
	Number    int32
	BlockType string // "FOLDER", "FILE", "POINTER"
	Content   interface{}
}

// readUsedBlocks lee todos los bloques que están marcados como utilizados
func readUsedBlocks(diskPath string, superblock *structs.Superblock, bitmap []byte) ([]blockInfo, error) {

	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var usedBlocks []blockInfo
	usedCount := 0

	for i := int32(0); i < superblock.S_blocks_count; i++ {
		if bitmap[i] == 1 { // Bloque en uso
			usedCount++

			// Calcular la posición correcta del bloque
			// Si S_block_start es 0, calcularlo manualmente
			var blockStartPos int32
			if superblock.S_block_start == 0 {
				// Calcular: inode_start + (inodes_count * inode_size)
				blockStartPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				if i == 0 {
				}
			} else {
				blockStartPos = superblock.S_block_start
			}

			blockPos := blockStartPos + (i * structs.BLOCK_SIZE)

			// Leer el bloque
			blockData := make([]byte, structs.BLOCK_SIZE)
			_, err := file.ReadAt(blockData, int64(blockPos))
			if err != nil {
				continue // Saltar bloques con error
			}

			// Determinar el tipo de bloque y leer su contenido
			blockType, content := determineBlockType(blockData)

			usedBlocks = append(usedBlocks, blockInfo{
				Number:    i,
				BlockType: blockType,
				Content:   content,
			})
		}
	}

	return usedBlocks, nil
}

// isValidFilename verifica si un nombre de archivo es válido
func isValidFilename(name string) bool {
	if name == "" || len(name) > 12 {
		return false
	}

	// Verificar que solo contenga caracteres imprimibles válidos
	for _, r := range name {
		if r < 32 || r > 126 {
			return false
		}
	}

	// Verificar que no contenga caracteres problemáticos
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, invalid := range invalidChars {
		if strings.Contains(name, invalid) {
			return false
		}
	}

	return true
}

// determineBlockType determina el tipo de bloque y extrae su contenido
func determineBlockType(blockData []byte) (string, interface{}) {

	// Debug específico para tipos de bloque
	hasNonZero := false
	for i := 0; i < 64 && i < len(blockData); i++ {
		if blockData[i] != 0 {
			hasNonZero = true
			break
		}
	}

	if !hasNonZero {
		return "FILE", blockData // Bloque vacío
	}

	// Intentar leer como FolderBlock PRIMERO
	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &folderBlock)
	if err == nil {
		// Verificar si tiene entradas válidas de directorio
		hasValidDirEntries := false
		validEntryCount := 0
		hasDirectoryMarkers := false

		for _, entry := range folderBlock.B_content {
			name := strings.TrimRight(string(entry.B_name[:]), "\x00")
			// Solo considerar entradas con nombres válidos y inodos razonables
			if name != "" && entry.B_inodo >= 0 && entry.B_inodo < 1000 {
				validEntryCount++
				// Verificar marcadores típicos de directorio
				if name == "." || name == ".." {
					hasDirectoryMarkers = true
					hasValidDirEntries = true
				} else if len(name) <= 12 && !strings.Contains(name, "\n") &&
					!strings.Contains(name, "\x00") && isValidFilename(name) {
					hasValidDirEntries = true
				}
			}
		}

		// Un bloque es de directorio si:
		// 1. Tiene marcadores de directorio (. y ..) OR
		// 2. Tiene al menos 2 entradas válidas con nombres razonables
		if (hasDirectoryMarkers && validEntryCount >= 2) ||
			(hasValidDirEntries && validEntryCount >= 2 && !hasDirectoryMarkers) {
			return "FOLDER", folderBlock
		}
	}

	// Intentar leer como PointerBlock
	var pointerBlock structs.PointerBlock
	buffer = bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err == nil {
		// Verificar si tiene punteros válidos
		validPointerCount := 0
		totalNonZeroPointers := 0

		for _, pointer := range pointerBlock.B_pointers {
			if pointer != 0 {
				totalNonZeroPointers++
				if pointer > 0 && pointer < 1000 {
					validPointerCount++
				}
			}
		}

		// Un bloque es de punteros si:
		// 1. Tiene al menos 2 punteros válidos OR
		// 2. Tiene muchos punteros no-cero (indicativo de datos estructurados como punteros)
		if validPointerCount >= 2 || (totalNonZeroPointers >= 3 && validPointerCount >= 1) {
			return "POINTER", pointerBlock
		}
	}

	// Intentar leer como FileBlock
	var fileBlock structs.FileBlock
	buffer = bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &fileBlock)
	if err == nil {
		content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
		if len(content) > 0 {
			printableChars := 0
			for _, r := range content {
				if (r >= 32 && r <= 126) || r == '\n' || r == '\t' || r == '\r' {
					printableChars++
				}
			}
			if float64(printableChars)/float64(len(content)) > 0.7 {
				return "FILE", fileBlock
			}
		}
	}

	return "FILE", blockData
} // generateBlockLabel genera la etiqueta para mostrar la información del bloque
func generateBlockLabel(index int, blockInfo blockInfo) string {
	var label string

	switch blockInfo.BlockType {
	case "FOLDER":
		folderBlock := blockInfo.Content.(structs.FolderBlock)
		label = fmt.Sprintf("BLOQUE %d\\n(CARPETA)", blockInfo.Number)

		// Mostrar TODAS las entradas del directorio, incluyendo las con inodo 0
		for i, entry := range folderBlock.B_content {
			name := strings.TrimRight(string(entry.B_name[:]), "\x00")
			if name != "" {
				label += fmt.Sprintf("\\n[%d] %s → %d", i, escapeLabel(name), entry.B_inodo)
			}
		}

	case "POINTER":
		pointerBlock := blockInfo.Content.(structs.PointerBlock)
		label = fmt.Sprintf("BLOQUE %d\\n(PUNTEROS)", blockInfo.Number)

		pointerCount := 0
		for i, pointer := range pointerBlock.B_pointers {
			if pointer > 0 {
				if pointerCount < 5 { // Mostrar solo los primeros 5 punteros
					label += fmt.Sprintf("\\n[%d] → %d", i, pointer)
				}
				pointerCount++
			}
		}

		if pointerCount > 5 {
			label += fmt.Sprintf("\\n... y %d más", pointerCount-5)
		} else if pointerCount == 0 {
			label += "\\nSin punteros válidos"
		}

	case "FILE":
		var content string

		// Intentar usar la estructura FileBlock primero
		if fileBlock, ok := blockInfo.Content.(structs.FileBlock); ok {
			// Extraer contenido del FileBlock
			content = strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
		} else if rawData, ok := blockInfo.Content.([]byte); ok {
			// Usar datos raw si no se pudo leer como FileBlock
			content = strings.TrimRight(string(rawData), "\x00")
		} else {
			content = "(contenido no disponible)"
		}

		// Limpiar y truncar el contenido para mostrar
		content = strings.TrimSpace(content)

		// Mostrar más contenido para archivos como users.txt
		maxLength := 100
		if len(content) > maxLength {
			content = content[:maxLength-3] + "..."
		}

		if content == "" {
			content = "(bloque vacío)"
		}

		// Escapar caracteres especiales para DOT
		content = escapeLabel(content)

		// Dividir contenido en líneas para mejor visualización vertical
		lines := strings.Split(content, "\\n")
		if len(lines) > 4 {
			lines = append(lines[:3], "...")
		}

		// Formatear como vertical
		label = fmt.Sprintf("BLOQUE %d\\n(ARCHIVO)", blockInfo.Number)
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				label += fmt.Sprintf("\\n%s", line)
			}
		}

	default:
		label = fmt.Sprintf("BLOQUE %d\\n(DESCONOCIDO)\\nTipo no identificado", blockInfo.Number)
	}

	return label
}

// DirectoryEntry representa una entrada en un directorio
type DirectoryEntry struct {
	Name     string
	InodeNum int32
}

// findInodeByPath encuentra el número de inodo de un archivo/directorio dado su path
func findInodeByPath(file *os.File, path string, superblock *structs.Superblock) (int32, error) {
	if path == "/" {
		return 0, nil // El inodo raíz es siempre 0
	}

	// Dividir el path en componentes
	components := strings.Split(strings.Trim(path, "/"), "/")
	currentInodeNum := int32(0) // Empezar desde la raíz

	for _, component := range components {
		if component == "" {
			continue
		}

		// Leer el inodo actual
		currentInode, err := readInodeAt(file, currentInodeNum, superblock.S_inode_start)
		if err != nil {
			return -1, fmt.Errorf("error leyendo inodo %d: %v", currentInodeNum, err)
		}

		// Verificar que sea un directorio
		if currentInode.I_type != 1 {
			return -1, fmt.Errorf("'%s' no es un directorio", component)
		}

		// Buscar el componente en el directorio actual
		found := false
		for i := 0; i < 12; i++ { // Bloques directos
			if currentInode.I_block[i] != 0 {
				blockBytes := make([]byte, structs.BLOCK_SIZE)
				blockPos := superblock.S_block_start + (currentInode.I_block[i] * structs.BLOCK_SIZE)
				_, err = file.ReadAt(blockBytes, int64(blockPos))
				if err != nil {
					continue
				}

				var folderBlock structs.FolderBlock
				buffer := bytes.NewReader(blockBytes)
				binary.Read(buffer, binary.LittleEndian, &folderBlock)

				for _, content := range folderBlock.B_content {
					entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
					if entryName == component {
						currentInodeNum = content.B_inodo
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}

		if !found {
			return -1, fmt.Errorf("no se encontró '%s' en el directorio", component)
		}
	}

	return currentInodeNum, nil
}

// readInodeAt lee un inodo desde una posición específica
func readInodeAt(file *os.File, inodeNum, inodeStart int32) (*structs.Inode, error) {
	inodeBytes := make([]byte, structs.INODE_SIZE)
	position := int64(inodeStart + (inodeNum * structs.INODE_SIZE))

	_, err := file.ReadAt(inodeBytes, position)
	if err != nil {
		return nil, fmt.Errorf("error leyendo inodo en posición %d: %v", position, err)
	}

	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	err = binary.Read(buffer, binary.LittleEndian, &inode)
	if err != nil {
		return nil, fmt.Errorf("error deserializando inodo: %v", err)
	}

	return &inode, nil
}

// getDirectoryEntries obtiene todas las entradas de un directorio
func getDirectoryEntries(file *os.File, dirInode *structs.Inode, superblock *structs.Superblock) ([]DirectoryEntry, error) {
	var entries []DirectoryEntry

	// Leer bloques directos del directorio
	for i := 0; i < 12; i++ { // Los primeros 12 son bloques directos
		if dirInode.I_block[i] != 0 {
			blockBytes := make([]byte, structs.BLOCK_SIZE)

			// Verificar si S_block_start es 0 y calcularlo manualmente si es necesario
			var blockPos int32
			if superblock.S_block_start == 0 {
				// Calcular: inode_start + (inodes_count * inode_size)
				blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				blockPos += (dirInode.I_block[i] * structs.BLOCK_SIZE)
			} else {
				blockPos = superblock.S_block_start + (dirInode.I_block[i] * structs.BLOCK_SIZE)
			}

			_, err := file.ReadAt(blockBytes, int64(blockPos))
			if err != nil {

				continue
			}

			var folderBlock structs.FolderBlock
			buffer := bytes.NewReader(blockBytes)
			err = binary.Read(buffer, binary.LittleEndian, &folderBlock)
			if err != nil {

				continue
			}

			// Procesar cada entrada en el bloque
			for j, content := range folderBlock.B_content {
				entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
				fmt.Printf("DEBUG getDirectoryEntries: Entrada %d: '%s' (inodo %d)\n", j, entryName, content.B_inodo)
				if entryName != "" && content.B_inodo >= 0 {
					entries = append(entries, DirectoryEntry{
						Name:     entryName,
						InodeNum: content.B_inodo,
					})
				}
			}
		}
	}

	return entries, nil
}

// formatPermissions convierte los permisos numéricos a formato rwx
func formatPermissions(perm int32) string {
	if perm == 0 {
		return "----------"
	}

	// Determinar tipo de archivo (primer carácter)
	fileType := "-"

	// Los siguientes 9 caracteres son rwx para owner, group, others
	perms := make([]byte, 9)

	// Owner permissions (bits 8-6)
	if perm&0400 != 0 {
		perms[0] = 'r'
	} else {
		perms[0] = '-'
	}
	if perm&0200 != 0 {
		perms[1] = 'w'
	} else {
		perms[1] = '-'
	}
	if perm&0100 != 0 {
		perms[2] = 'x'
	} else {
		perms[2] = '-'
	}

	// Group permissions (bits 5-3)
	if perm&0040 != 0 {
		perms[3] = 'r'
	} else {
		perms[3] = '-'
	}
	if perm&0020 != 0 {
		perms[4] = 'w'
	} else {
		perms[4] = '-'
	}
	if perm&0010 != 0 {
		perms[5] = 'x'
	} else {
		perms[5] = '-'
	}

	// Others permissions (bits 2-0)
	if perm&0004 != 0 {
		perms[6] = 'r'
	} else {
		perms[6] = '-'
	}
	if perm&0002 != 0 {
		perms[7] = 'w'
	} else {
		perms[7] = '-'
	}
	if perm&0001 != 0 {
		perms[8] = 'x'
	} else {
		perms[8] = '-'
	}

	return fileType + string(perms)
}

// UserInfo estructura para almacenar información de usuario
type UserInfo struct {
	Username string
	Group    string
	UID      int32
	GID      int32
}

// readUsersFile lee el archivo users.txt del sistema de archivos y extrae información de usuarios
func readUsersFile(file *os.File, superblock *structs.Superblock) (map[int32]UserInfo, error) {
	usersMap := make(map[int32]UserInfo)

	// Buscar el archivo users.txt de manera recursiva en el sistema de archivos
	usersInodeNum, err := findFileRecursively(file, "users.txt", 0, superblock)
	if err != nil {
		return usersMap, fmt.Errorf("no se encontró users.txt: %v", err)
	}

	// Leer el inodo del archivo users.txt
	usersInode, err := readInodeAt(file, usersInodeNum, superblock.S_inode_start)
	if err != nil {
		return usersMap, fmt.Errorf("error leyendo inodo de users.txt: %v", err)
	}

	// Leer el contenido del archivo users.txt
	content, err := readFileContent(file, usersInode, superblock)
	if err != nil {
		return usersMap, fmt.Errorf("error leyendo contenido de users.txt: %v", err)
	}

	// Parsear el contenido del archivo users.txt
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 5 {
			// Formato: ID,tipo,grupo,nombre,password
			userType := strings.TrimSpace(parts[1])
			if userType == "U" { // Es un usuario
				if uid, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					username := strings.TrimSpace(parts[3])
					groupName := strings.TrimSpace(parts[2])

					usersMap[int32(uid)] = UserInfo{
						Username: username,
						Group:    groupName,
						UID:      int32(uid),
						GID:      0, // Se puede calcular después si es necesario
					}
				}
			}
		}
	}

	return usersMap, nil
}

// findFileInRoot busca un archivo específico en el directorio raíz
func findFileInRoot(file *os.File, filename string, superblock *structs.Superblock) (int32, error) {
	// Leer el inodo raíz (siempre es el inodo 0)
	rootInode, err := readInodeAt(file, 0, superblock.S_inode_start)
	if err != nil {
		return -1, fmt.Errorf("error leyendo inodo raíz: %v", err)
	}

	// Buscar en los bloques del directorio raíz
	for i := 0; i < 12; i++ { // Bloques directos
		if rootInode.I_block[i] != 0 {
			blockBytes := make([]byte, structs.BLOCK_SIZE)

			var blockPos int32
			if superblock.S_block_start == 0 {
				blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				blockPos += (rootInode.I_block[i] * structs.BLOCK_SIZE)
			} else {
				blockPos = superblock.S_block_start + (rootInode.I_block[i] * structs.BLOCK_SIZE)
			}

			_, err = file.ReadAt(blockBytes, int64(blockPos))
			if err != nil {
				continue
			}

			var folderBlock structs.FolderBlock
			buffer := bytes.NewReader(blockBytes)
			binary.Read(buffer, binary.LittleEndian, &folderBlock)

			for _, content := range folderBlock.B_content {
				entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
				if entryName == filename {
					return content.B_inodo, nil
				}
			}
		}
	}

	return -1, fmt.Errorf("archivo '%s' no encontrado en el directorio raíz", filename)
}

// findFileRecursively busca un archivo de manera recursiva en el sistema de archivos
func findFileRecursively(file *os.File, filename string, currentInodeNum int32, superblock *structs.Superblock) (int32, error) {
	// Leer el inodo actual
	currentInode, err := readInodeAt(file, currentInodeNum, superblock.S_inode_start)
	if err != nil {
		return -1, fmt.Errorf("error leyendo inodo %d: %v", currentInodeNum, err)
	}

	// Si es un archivo, verificar si es el que buscamos
	if currentInode.I_type == 0 {
		// Es un archivo, no podemos buscar dentro de él
		return -1, fmt.Errorf("no es un directorio")
	}

	// Si es un directorio, buscar en sus entradas
	if currentInode.I_type == 1 {
		// Buscar en los bloques del directorio
		for i := 0; i < 12; i++ { // Bloques directos
			if currentInode.I_block[i] != 0 {
				blockBytes := make([]byte, structs.BLOCK_SIZE)

				var blockPos int32
				if superblock.S_block_start == 0 {
					blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
					blockPos += (currentInode.I_block[i] * structs.BLOCK_SIZE)
				} else {
					blockPos = superblock.S_block_start + (currentInode.I_block[i] * structs.BLOCK_SIZE)
				}

				_, err = file.ReadAt(blockBytes, int64(blockPos))
				if err != nil {
					continue
				}

				var folderBlock structs.FolderBlock
				buffer := bytes.NewReader(blockBytes)
				binary.Read(buffer, binary.LittleEndian, &folderBlock)

				for _, content := range folderBlock.B_content {
					entryName := strings.TrimRight(string(content.B_name[:]), "\x00")

					// Saltar entradas vacías
					if entryName == "" {
						continue
					}

					// Si encontramos el archivo que buscamos
					if entryName == filename {
						return content.B_inodo, nil
					}

					// Saltar entradas "." y ".." para evitar recursión infinita
					if entryName == "." || entryName == ".." {
						continue
					}

					// Si es un directorio, buscar recursivamente (evitar recursión infinita)
					if content.B_inodo != currentInodeNum && content.B_inodo > 0 {
						entryInode, err := readInodeAt(file, content.B_inodo, superblock.S_inode_start)
						if err == nil && entryInode.I_type == 1 {
							// Es un directorio, buscar recursivamente
							result, err := findFileRecursively(file, filename, content.B_inodo, superblock)
							if err == nil {
								return result, nil
							}
						}
					}
				}
			}
		}
	}

	return -1, fmt.Errorf("archivo '%s' no encontrado", filename)
} // readFileContent lee el contenido completo de un archivo dado su inodo
func readFileContent(file *os.File, fileInode *structs.Inode, superblock *structs.Superblock) (string, error) {
	var content strings.Builder

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		if fileInode.I_block[i] != 0 {
			blockBytes := make([]byte, structs.BLOCK_SIZE)

			var blockPos int32
			if superblock.S_block_start == 0 {
				blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				blockPos += (fileInode.I_block[i] * structs.BLOCK_SIZE)
			} else {
				blockPos = superblock.S_block_start + (fileInode.I_block[i] * structs.BLOCK_SIZE)
			}

			_, err := file.ReadAt(blockBytes, int64(blockPos))
			if err != nil {
				continue
			}

			// Leer como FileBlock
			var fileBlock structs.FileBlock
			buffer := bytes.NewReader(blockBytes)
			err = binary.Read(buffer, binary.LittleEndian, &fileBlock)
			if err != nil {
				continue
			}

			blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			content.WriteString(blockContent)
		}
	}

	return content.String(), nil
}
