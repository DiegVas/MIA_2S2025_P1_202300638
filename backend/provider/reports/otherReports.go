package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
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
	for i, inode := range usedInodes {
		inodeLabel := generateInodeLabel(i, inode)
		dot.WriteString(fmt.Sprintf("    inode_%d [label=\"%s\", fillcolor=lightyellow];\n", i, inodeLabel))
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
		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		fmt.Printf("[DEBUG] Partición %d: Name='%s', Status=%d, Type=%d, Fit=%s, Start=%d, Size=%d\n",
			i, partName, partition.Part_status, partition.Part_type, partition.Part_fit, partition.Part_start, partition.Part_size)
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

	fmt.Printf("[DEBUG] Superbloque leído:\n")
	fmt.Printf("[DEBUG]   Magic: 0x%X\n", superblock.S_magic)
	fmt.Printf("[DEBUG]   Inodos: %d, Bloques: %d\n", superblock.S_inodes_count, superblock.S_blocks_count)
	fmt.Printf("[DEBUG]   BM Inode start: %d\n", superblock.S_bm_inode_start)
	fmt.Printf("[DEBUG]   BM Block start: %d\n", superblock.S_bm_block_start)
	fmt.Printf("[DEBUG]   Inode start: %d\n", superblock.S_inode_start)
	fmt.Printf("[DEBUG]   Block start: %d\n", superblock.S_block_start)

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

	// Formatear los bloques (solo mostrar los primeros 5 para no sobrecargar)
	var blocks []string
	for i := 0; i < 5 && i < len(inode.I_block); i++ {
		if inode.I_block[i] != 0 {
			blocks = append(blocks, fmt.Sprintf("%d", inode.I_block[i]))
		}
	}
	blocksStr := strings.Join(blocks, ", ")
	if len(inode.I_block) > 5 {
		blocksStr += "..."
	}

	// Generar la etiqueta con toda la información
	label := fmt.Sprintf("INODO %d|"+
		"{UID: %d|GID: %d}|"+
		"{Tamaño: %d bytes|Tipo: %s}|"+
		"{Permisos: %o|Bloques: [%s]}|"+
		"{A-time: %s}|"+
		"{C-time: %s}|"+
		"{M-time: %s}",
		inodeInfo.Number,
		inode.I_uid, inode.I_gid,
		inode.I_size, inodeType,
		inode.I_perm, blocksStr,
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
	fmt.Printf("[DEBUG] Iniciando generación de reporte de bloques para partición: %s\n", partitionID)

	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}
	fmt.Printf("[DEBUG] Partición encontrada: %s\n", mountedPartition.Path)

	// Leer el superbloque para obtener información del sistema de archivos
	superblock, err := readSuperblock(mountedPartition.Path, partitionID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}
	fmt.Printf("[DEBUG] Superbloque leído. Bloques totales: %d, inicio bloques: %d\n",
		superblock.S_blocks_count, superblock.S_block_start)

	// Leer bitmap de bloques para identificar cuáles están en uso
	blockBitmap, err := readBlockBitmap(mountedPartition.Path, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}
	fmt.Printf("[DEBUG] Bitmap de bloques leído\n")

	// Leer todos los bloques utilizados
	usedBlocks, err := readUsedBlocks(mountedPartition.Path, superblock, blockBitmap)
	if err != nil {
		return "", fmt.Errorf("error leyendo bloques: %v", err)
	}

	if len(usedBlocks) == 0 {
		fmt.Printf("[DEBUG] No se encontraron bloques utilizados\n")
		return "", fmt.Errorf("no se encontraron bloques utilizados")
	}

	fmt.Printf("[DEBUG] Generando contenido DOT para %d bloques\n", len(usedBlocks))

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

// generateBMInodeReport genera el reporte del bitmap de inodos
func generateBMInodeReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph BM_Inode_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"BM Inode Report\" [label=\"Bitmap de Inodos\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("}\n")
	return dot.String(), nil
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

// generateTreeReport genera el reporte del árbol de directorios
func generateTreeReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Tree_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=box];\n")
	dot.WriteString("    root [label=\"/\"];\n")
	dot.WriteString("    \"Tree Report\" [label=\"Árbol de Directorios\\nPartición: " + partitionID + "\"];\n")
	dot.WriteString("    \"Tree Report\" -> root;\n")
	dot.WriteString("}\n")
	return dot.String(), nil
}

// generateSuperblockReport genera el reporte del superbloque
func generateSuperblockReport(partitionID string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph Superblock_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString("    \"Superblock Report\" [label=\"Superbloque\\nPartición: " + partitionID + "\"];\n")
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

// generateLSReport genera el reporte de listado de directorio
func generateLSReport(partitionID, dirPath string) (string, error) {
	var dot strings.Builder
	dot.WriteString("digraph LS_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=record];\n")
	dot.WriteString(fmt.Sprintf("    \"LS Report\" [label=\"Listado: %s\\nPartición: %s\"];\n",
		dirPath, partitionID))
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
	fmt.Printf("[DEBUG] Calculando posición bitmap bloques: %d + %d = %d\n",
		superblock.S_bm_inode_start, superblock.S_inodes_count, bmBlockPos)

	bitmap := make([]byte, superblock.S_blocks_count)
	_, err = file.ReadAt(bitmap, int64(bmBlockPos))
	if err != nil {
		return nil, err
	}

	// Mostrar los primeros bits del bitmap para debug
	fmt.Printf("[DEBUG] Primeros 10 bits del bitmap de bloques: ")
	for i := 0; i < 10 && i < len(bitmap); i++ {
		fmt.Printf("%d ", bitmap[i])
	}
	fmt.Printf("\n")

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
	fmt.Printf("[DEBUG] Leyendo bloques utilizados. Total bloques: %d\n", superblock.S_blocks_count)

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
					fmt.Printf("[DEBUG] S_block_start era 0, calculado como: %d\n", blockStartPos)
				}
			} else {
				blockStartPos = superblock.S_block_start
			}

			blockPos := blockStartPos + (i * structs.BLOCK_SIZE)
			fmt.Printf("[DEBUG] Bloque %d en uso, posición: %d\n", i, blockPos)

			// Leer el bloque
			blockData := make([]byte, structs.BLOCK_SIZE)
			_, err := file.ReadAt(blockData, int64(blockPos))
			if err != nil {
				fmt.Printf("[DEBUG] Error leyendo bloque %d: %v\n", i, err)
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

	fmt.Printf("[DEBUG] Total bloques utilizados encontrados: %d\n", usedCount)
	return usedBlocks, nil
}

// determineBlockType determina el tipo de bloque y extrae su contenido
func determineBlockType(blockData []byte) (string, interface{}) {
	fmt.Printf("[DEBUG] Analizando bloque de %d bytes\n", len(blockData))

	// Intentar leer como FolderBlock
	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &folderBlock)
	if err == nil {
		// Verificar si tiene entradas válidas de directorio
		hasValidEntries := false
		for _, entry := range folderBlock.B_content {
			name := strings.TrimRight(string(entry.B_name[:]), "\x00")
			if name != "" && entry.B_inodo > 0 {
				hasValidEntries = true
				fmt.Printf("[DEBUG] Entrada de directorio encontrada: %s -> inodo %d\n", name, entry.B_inodo)
				break
			}
		}
		if hasValidEntries {
			fmt.Printf("[DEBUG] Bloque identificado como FOLDER\n")
			return "FOLDER", folderBlock
		}
	}

	// Intentar leer como PointerBlock
	var pointerBlock structs.PointerBlock
	buffer = bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err == nil {
		// Verificar si tiene punteros válidos
		hasValidPointers := false
		for _, pointer := range pointerBlock.B_pointers {
			if pointer > 0 {
				hasValidPointers = true
				fmt.Printf("[DEBUG] Puntero válido encontrado: %d\n", pointer)
				break
			}
		}
		if hasValidPointers {
			fmt.Printf("[DEBUG] Bloque identificado como POINTER\n")
			return "POINTER", pointerBlock
		}
	}

	// Leer como FileBlock (estructura de 64 bytes)
	var fileBlock structs.FileBlock
	buffer = bytes.NewReader(blockData)
	err = binary.Read(buffer, binary.LittleEndian, &fileBlock)
	if err == nil {
		content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
		fmt.Printf("[DEBUG] Bloque identificado como FILE, contenido: '%.50s'\n", content)
		return "FILE", fileBlock
	}

	// Si no se puede leer como FileBlock, usar los datos raw
	fmt.Printf("[DEBUG] Bloque no identificado, usando datos raw\n")
	return "FILE", blockData
}

// generateBlockLabel genera la etiqueta para mostrar la información del bloque
func generateBlockLabel(index int, blockInfo blockInfo) string {
	var label string

	switch blockInfo.BlockType {
	case "FOLDER":
		folderBlock := blockInfo.Content.(structs.FolderBlock)
		label = fmt.Sprintf("BLOQUE %d (CARPETA)|", blockInfo.Number)

		entryCount := 0
		for _, entry := range folderBlock.B_content {
			name := strings.TrimRight(string(entry.B_name[:]), "\x00")
			if name != "" && entry.B_inodo > 0 {
				if entryCount < 3 { // Mostrar solo las primeras 3 entradas
					label += fmt.Sprintf("{%s → inodo %d}|", escapeLabel(name), entry.B_inodo)
				}
				entryCount++
			}
		}

		if entryCount > 3 {
			label += fmt.Sprintf("{... y %d más}", entryCount-3)
		} else if entryCount == 0 {
			label += "{Carpeta vacía}"
		} else {
			label = strings.TrimSuffix(label, "|") // Quitar el último separador
		}

	case "POINTER":
		pointerBlock := blockInfo.Content.(structs.PointerBlock)
		label = fmt.Sprintf("BLOQUE %d (PUNTEROS)|", blockInfo.Number)

		pointerCount := 0
		for i, pointer := range pointerBlock.B_pointers {
			if pointer > 0 {
				if pointerCount < 5 { // Mostrar solo los primeros 5 punteros
					label += fmt.Sprintf("{[%d] → %d}|", i, pointer)
				}
				pointerCount++
			}
		}

		if pointerCount > 5 {
			label += fmt.Sprintf("{... y %d más}", pointerCount-5)
		} else if pointerCount == 0 {
			label += "{Sin punteros válidos}"
		} else {
			label = strings.TrimSuffix(label, "|")
		}

	case "FILE":
		fmt.Printf("[DEBUG] Procesando bloque %d como FILE\n", blockInfo.Number)
		var content string

		// Intentar usar la estructura FileBlock primero
		if fileBlock, ok := blockInfo.Content.(structs.FileBlock); ok {
			// Extraer contenido del FileBlock
			content = strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			fmt.Printf("[DEBUG] Contenido de FileBlock (primeros 50 chars): '%.50s'\n", content)
		} else if rawData, ok := blockInfo.Content.([]byte); ok {
			// Usar datos raw si no se pudo leer como FileBlock
			content = strings.TrimRight(string(rawData), "\x00")
			fmt.Printf("[DEBUG] Contenido de datos raw (primeros 50 chars): '%.50s'\n", content)
		} else {
			content = "(contenido no disponible)"
			fmt.Printf("[DEBUG] No se pudo extraer contenido del bloque\n")
		}

		// Limpiar y truncar el contenido para mostrar
		content = strings.TrimSpace(content)
		fmt.Printf("[DEBUG] Contenido después de limpiar: '%.50s'\n", content)

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

		// Dividir contenido en líneas para mejor visualización
		lines := strings.Split(content, "\\n")
		if len(lines) > 5 {
			lines = append(lines[:4], "...")
		}

		contentDisplay := strings.Join(lines, "\\l")

		label = fmt.Sprintf("BLOQUE %d (ARCHIVO)|{Contenido:\\l%s}",
			blockInfo.Number, contentDisplay)

	default:
		label = fmt.Sprintf("BLOQUE %d (DESCONOCIDO)|{Tipo no identificado}", blockInfo.Number)
	}

	return label
}
