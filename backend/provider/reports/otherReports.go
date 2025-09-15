package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

// generateInodeLabel genera la etiqueta para mostrar la información del inodo
func generateInodeLabel(_ int, inodeInfo inodeInfo) string {

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
		"{Tamano: %d bytes | Permisos: %s}|"+
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

// inodeInfo estructura auxiliar para almacenar información del inodo
type inodeInfo struct {
	Number int32
	Inode  structs.Inode
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

// generateDirectoryBlockToInodeConnections genera conexiones desde bloques de directorio hacia inodos
func generateDirectoryBlockToInodeConnections(file *os.File, superblock *structs.Superblock, inodeBitmap, blockBitmap []byte, dot *strings.Builder) error {

	// Recorrer bloques utilizados que son de tipo directorio
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		if i < int32(len(blockBitmap)) && blockBitmap[i] == 1 {
			// Leer el bloque para determinar si es un directorio
			blockData, blockType, err := readBlockAt(file, i, superblock)
			if err != nil {
				continue
			}

			// Solo procesar bloques de directorio
			if blockType == "Directorio" {

				err = generateDirectoryEntryConnections(i, blockData, inodeBitmap, dot)
				if err != nil {
					fmt.Printf("DEBUG TREE: Error procesando entradas del bloque %d: %v\n", i, err)
				}
			}
		}
	}

	return nil
}

// generateDirectoryEntryConnections genera conexiones desde entradas de directorio hacia inodos
func generateDirectoryEntryConnections(blockNum int32, blockData []byte, inodeBitmap []byte, dot *strings.Builder) error {
	var folderBlock structs.FolderBlock
	buffer := bytes.NewReader(blockData)
	err := binary.Read(buffer, binary.LittleEndian, &folderBlock)
	if err != nil {
		return err
	}

	for i, content := range folderBlock.B_content {
		entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
		inodeNum := content.B_inodo

		// Solo crear conexión si la entrada no está vacía y el inodo existe
		if entryName != "" && inodeNum >= 0 && inodeNum < int32(len(inodeBitmap)) && inodeBitmap[inodeNum] == 1 {
			dot.WriteString(fmt.Sprintf("    block_%d:entry_%d -> inode_%d;\n", blockNum, i, inodeNum))
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
	table.WriteString(fmt.Sprintf("<tr><td>Tamano</td><td>%d bytes</td></tr>", inode.I_size))
	table.WriteString(fmt.Sprintf("<tr><td>Permisos</td><td>%o</td></tr>", inode.I_perm))
	table.WriteString(fmt.Sprintf("<tr><td>A-time</td><td>%s</td></tr>", atime))
	table.WriteString(fmt.Sprintf("<tr><td>C-time</td><td>%s</td></tr>", ctime))
	table.WriteString(fmt.Sprintf("<tr><td>M-time</td><td>%s</td></tr>", mtime))

	// Mostrar bloques directos e indirectos
	for i := 0; i < 15; i++ {
		if inode.I_block[i] != 0 {
			blockType := "Directo"
			switch i {
			case 12:
				blockType = "Indirecto Simple"
			case 13:
				blockType = "Indirecto Doble"
			case 14:
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
			fmt.Printf("DEBUG TREE: Error generando tabla de directorio para bloque %d: %v\n", blockNum, err)
			table.WriteString(fmt.Sprintf("<tr><td>Error</td><td>%s</td></tr>", err.Error()))
		}
	case "Archivo":
		generateFileBlockTable(&table, blockData)
	case "Punteros":
		err := generatePointerBlockTable(&table, blockData)
		if err != nil {
			fmt.Printf("DEBUG TREE: Error generando tabla de punteros para bloque %d: %v\n", blockNum, err)
			table.WriteString(fmt.Sprintf("<tr><td>Error</td><td>%s</td></tr>", err.Error()))
		}
	default:
		table.WriteString("<tr><td>Contenido</td><td>Datos binarios</td></tr>")
	}

	table.WriteString("</table>")
	result := table.String()

	return result
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

			// Escapar caracteres especiales en nombres de archivos/directorios
			entryName = escapeHTMLForDot(entryName)

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

	// Si el contenido está vacío, mostrar mensaje
	if content == "" {
		content = "[vacio]"

	}

	// Escapar caracteres especiales para HTML en DOT
	content = escapeHTMLForDot(content)

	table.WriteString(fmt.Sprintf("<tr><td>Contenido</td><td>%s</td></tr>", content))
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

// blockInfo estructura auxiliar para almacenar información del bloque
type blockInfo struct {
	Number    int32
	BlockType string // "FOLDER", "FILE", "POINTER"
	Content   interface{}
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
func generateBlockLabel(_ int, blockInfo blockInfo) string {
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
			if currentInode.I_block[i] != 0 { // Verificar que el bloque sea válido

				blockBytes := make([]byte, structs.BLOCK_SIZE)
				// Calcular la posición correcta del bloque dentro de la partición
				var blockPos int32
				if superblock.S_block_start == 0 {
					// Calcular: inode_start + (inodes_count * inode_size)
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
					if entryName != "" {
					}
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
				fmt.Printf("DEBUG LS BLOCK %d ENTRY %d: Name='%s', Inode=%d\n", dirInode.I_block[i], j, entryName, content.B_inodo)
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

// UserInfo estructura para almacenar información de usuario
type UserInfo struct {
	Username string
	Group    string
	UID      int32
	GID      int32
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
}
