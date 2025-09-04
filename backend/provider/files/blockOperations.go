package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/structs"
	"os"
)

func FindFreeBlocks(file *os.File, bmBlockPos, blockCount, blocksNeeded int32) ([]int32, error) {

	bitmap := make([]byte, blockCount)
	_, err := file.ReadAt(bitmap, int64(bmBlockPos))
	if err != nil {
		fmt.Printf("❌ ERROR leyendo bitmap: %v\n", err)
		return nil, err
	}

	// Debug: mostrar estado del bitmap
	usedBlocks := 0
	for i := int32(0); i < blockCount; i++ {
		if bitmap[i] == 1 {
			usedBlocks++
		}
	}

	var freeBlocks []int32
	for i := int32(2); i < blockCount && int32(len(freeBlocks)) < blocksNeeded; i++ {
		if bitmap[i] == 0 {
			freeBlocks = append(freeBlocks, i)
			fmt.Printf("🔍 DEBUG FindFreeBlocks: Bloque libre encontrado: %d\n", i)

			// Marcar inmediatamente como usado en el bitmap para evitar reutilización
			bitmap[i] = 1
			_, err = file.WriteAt([]byte{1}, int64(bmBlockPos+i))
			if err != nil {
				fmt.Printf("❌ ERROR marcando bloque %d como usado: %v\n", i, err)
				return nil, err
			}
			fmt.Printf("🔍 DEBUG: Bloque %d marcado inmediatamente como usado\n", i)
		}
	}

	fmt.Printf("🔍 DEBUG FindFreeBlocks: Bloques libres encontrados: %v\n", freeBlocks)

	if int32(len(freeBlocks)) < blocksNeeded {
		fmt.Printf("❌ ERROR: no hay suficientes bloques libres (encontrados=%d, necesarios=%d)\n", len(freeBlocks), blocksNeeded)
		return nil, errors.New("no hay suficientes bloques libres disponibles")
	}

	return freeBlocks, nil
}

func WriteContentToBlocks(file *os.File, content string, blocks []int32, blockPos int32) error {
	contentBytes := []byte(content)
	blockIndex := 0

	for blockIndex < len(blocks) && (blockIndex*int(structs.BLOCK_SIZE)) < len(contentBytes) {
		blockData := make([]byte, structs.BLOCK_SIZE)

		start := blockIndex * int(structs.BLOCK_SIZE)
		end := start + int(structs.BLOCK_SIZE)
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		if start < len(contentBytes) {
			copy(blockData, contentBytes[start:end])
		}

		blockNum := blocks[blockIndex]
		_, err := file.WriteAt(blockData, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return fmt.Errorf("error escribiendo en bloque %d: %v", blockNum, err)
		}

		blockIndex++
	}
	return nil
}

// WriteContentToInodeBlocks escribe contenido usando la estructura de bloques del inodo
func WriteContentToInodeBlocks(file *os.File, content string, inode *structs.Inode, blockPos int32) error {
	contentBytes := []byte(content)
	writtenBytes := 0

	fmt.Printf("📝 DEBUG WriteContentToInodeBlocks: Escribiendo %d bytes\n", len(contentBytes))
	fmt.Printf("📝 DEBUG: Bloques del inodo: %v\n", inode.I_block)

	// 1. Escribir en bloques directos (slots 0-11)
	fmt.Printf("📝 DEBUG: Escribiendo en bloques directos\n")
	for i := 0; i < 12 && inode.I_block[i] != 0 && writtenBytes < len(contentBytes); i++ {
		blockData := make([]byte, structs.BLOCK_SIZE)

		// Crear estructura FileBlock
		var fileBlock structs.FileBlock

		start := writtenBytes
		end := start + structs.FILE_CONTENT_SIZE
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		if start < len(contentBytes) {
			copy(fileBlock.B_content[:], contentBytes[start:end])
		}

		// Serializar FileBlock a bytes
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.LittleEndian, fileBlock)
		copy(blockData, buffer.Bytes())

		blockNum := inode.I_block[i]
		position := int64(blockPos + (blockNum * structs.BLOCK_SIZE))
		bytesToWrite := end - start

		fmt.Printf("📝 DEBUG: Bloque directo %d -> Escribiendo %d bytes útiles en posición %d\n", blockNum, bytesToWrite, position)

		_, err := file.WriteAt(blockData, position)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque directo %d: %v", blockNum, err)
		}

		writtenBytes += structs.FILE_CONTENT_SIZE
	}

	fmt.Printf("📝 DEBUG: Bytes escritos en bloques directos: %d\n", writtenBytes)

	// 2. Escribir en bloques indirectos simples (slot 12)
	if inode.I_block[12] != 0 && writtenBytes < len(contentBytes) {
		fmt.Printf("📝 DEBUG: Escribiendo en bloque indirecto simple (slot 12), bloque: %d\n", inode.I_block[12])
		err := writeToIndirectBlock(file, contentBytes[writtenBytes:], inode.I_block[12], blockPos)
		if err != nil {
			return err
		}
		// Calcular cuántos bytes se escribieron en el bloque indirecto
		remainingBytes := len(contentBytes) - writtenBytes
		indirectBytes := 16 * structs.FILE_CONTENT_SIZE // Máximo que puede almacenar un bloque indirecto
		if remainingBytes < indirectBytes {
			writtenBytes += remainingBytes
		} else {
			writtenBytes += indirectBytes
		}
		fmt.Printf("📝 DEBUG: Bytes escritos después de indirecto simple: %d\n", writtenBytes)
	}

	// 3. Escribir en bloques doblemente indirectos (slot 13)
	if inode.I_block[13] != 0 && writtenBytes < len(contentBytes) {
		fmt.Printf("📝 DEBUG: Escribiendo en bloque doblemente indirecto (slot 13), bloque: %d\n", inode.I_block[13])
		err := writeToDoubleIndirectBlock(file, contentBytes[writtenBytes:], inode.I_block[13], blockPos)
		if err != nil {
			return err
		}
	}

	fmt.Printf("📝 DEBUG: Escritura completada. Total bytes escritos: %d de %d\n", writtenBytes, len(contentBytes))
	return nil
}

// writeToIndirectBlock escribe contenido en un bloque indirecto simple
func writeToIndirectBlock(file *os.File, content []byte, ptrBlockNum int32, blockPos int32) error {
	fmt.Printf("📝 DEBUG writeToIndirectBlock: Bloque de punteros %d, contenido: %d bytes\n", ptrBlockNum, len(content))

	// Validar que el número de bloque sea válido
	if ptrBlockNum < 0 {
		return fmt.Errorf("número de bloque de punteros inválido (negativo): %d", ptrBlockNum)
	}

	// Leer el bloque de punteros
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	ptrPosition := int64(blockPos + (ptrBlockNum * structs.BLOCK_SIZE))

	fmt.Printf("📝 DEBUG: Leyendo bloque de punteros en posición: %d\n", ptrPosition)

	_, err := file.ReadAt(blockBytes, ptrPosition)
	if err != nil {
		return fmt.Errorf("error leyendo bloque de punteros %d: %v", ptrBlockNum, err)
	}

	// Debug: Mostrar los primeros bytes del bloque leído
	fmt.Printf("📝 DEBUG: Primeros 64 bytes del bloque leído: %v\n", blockBytes[:64])

	var pointerBlock structs.PointerBlock
	buffer := bytes.NewReader(blockBytes)
	err = binary.Read(buffer, binary.LittleEndian, &pointerBlock)
	if err != nil {
		return fmt.Errorf("error deserializando bloque de punteros: %v", err)
	}

	fmt.Printf("📝 DEBUG: Punteros leídos: %v\n", pointerBlock.B_pointers)

	// Debug: Verificar si hay punteros inválidos
	for i, ptr := range pointerBlock.B_pointers {
		if ptr != 0 && (ptr < 0 || ptr > 100000) {
			fmt.Printf("📝 DEBUG ERROR: Puntero inválido en posición %d: %d\n", i, ptr)
		}
	}

	writtenBytes := 0

	// Escribir en cada bloque apuntado
	for i := 0; i < 16 && pointerBlock.B_pointers[i] != 0 && writtenBytes < len(content); i++ {
		dataBlockNum := pointerBlock.B_pointers[i]

		// Validar que el número de bloque sea válido
		if dataBlockNum < 0 {
			return fmt.Errorf("número de bloque de datos inválido (negativo): %d en posición %d", dataBlockNum, i)
		}

		// Verificar que el bloque no sea extremadamente grande (posible corrupción)
		maxReasonableBlock := int32(100000) // Ajustar según el tamaño de tu disco
		if dataBlockNum > maxReasonableBlock {
			return fmt.Errorf("número de bloque sospechosamente grande: %d en posición %d (max razonable: %d)", dataBlockNum, i, maxReasonableBlock)
		}

		blockData := make([]byte, structs.BLOCK_SIZE)

		// Crear estructura FileBlock
		var fileBlock structs.FileBlock

		start := writtenBytes
		end := start + structs.FILE_CONTENT_SIZE
		if end > len(content) {
			end = len(content)
		}

		if start < len(content) {
			copy(fileBlock.B_content[:], content[start:end])
		}

		// Serializar FileBlock a bytes
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.LittleEndian, fileBlock)
		copy(blockData, buffer.Bytes())

		dataPosition := int64(blockPos + (dataBlockNum * structs.BLOCK_SIZE))
		bytesToWrite := end - start

		// Validar que la posición no sea negativa
		if dataPosition < 0 {
			return fmt.Errorf("posición de escritura negativa: %d para bloque %d", dataPosition, dataBlockNum)
		}

		fmt.Printf("📝 DEBUG: Escribiendo %d bytes útiles en bloque indirecto %d (posición %d)\n", bytesToWrite, dataBlockNum, dataPosition)

		_, err := file.WriteAt(blockData, dataPosition)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque indirecto %d en posición %d: %v", dataBlockNum, dataPosition, err)
		}

		writtenBytes += structs.FILE_CONTENT_SIZE
	}

	fmt.Printf("📝 DEBUG writeToIndirectBlock: Escritura completada, %d bytes escritos\n", writtenBytes)
	return nil
} // writeToDoubleIndirectBlock escribe contenido en un bloque doblemente indirecto
func writeToDoubleIndirectBlock(file *os.File, content []byte, firstLevelBlockNum int32, blockPos int32) error {
	// Leer el bloque de primer nivel
	blockBytes := make([]byte, structs.BLOCK_SIZE)
	_, err := file.ReadAt(blockBytes, int64(blockPos+(firstLevelBlockNum*structs.BLOCK_SIZE)))
	if err != nil {
		return err
	}

	var firstLevelPointers structs.PointerBlock
	buffer := bytes.NewReader(blockBytes)
	binary.Read(buffer, binary.LittleEndian, &firstLevelPointers)

	writtenBytes := 0

	// Para cada entrada en el primer nivel
	for i := 0; i < 16 && firstLevelPointers.B_pointers[i] != 0 && writtenBytes < len(content); i++ {
		remainingContent := content[writtenBytes:]
		err := writeToIndirectBlock(file, remainingContent, firstLevelPointers.B_pointers[i], blockPos)
		if err != nil {
			return err
		}

		// Calcular cuántos bytes se escribieron
		bytesToWrite := 16 * structs.FILE_CONTENT_SIZE // Máximo por bloque indirecto
		if len(remainingContent) < bytesToWrite {
			writtenBytes += len(remainingContent)
		} else {
			writtenBytes += bytesToWrite
		}
	}

	return nil
}

func MarkBlockAsUsed(file *os.File, bmBlockPos, blockNum int32) error {
	_, err := file.WriteAt([]byte{1}, int64(bmBlockPos+blockNum))
	return err
}
