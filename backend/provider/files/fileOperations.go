package files

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"time"
)

func minLocal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CreateNewFile(file *os.File, fileName, content string, size int32, user *structs.User,
	parentInodeNum, inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount int32) error {

	fmt.Printf("📁 DEBUG CreateNewFile: Creando archivo '%s' con tamaño %d bytes\n", fileName, size)
	fmt.Printf("📁 DEBUG CreateNewFile: Contenido recibido tamaño %d, primeros 100 chars: %q\n", len(content), content[:minLocal(100, len(content))])

	// ! Encontrar un inodo libre
	freeInodeNum, err := FindFreeInode(file, bmInodePos, inodeCount)
	if err != nil {
		return err
	}
	fmt.Printf("📁 DEBUG: Inodo libre encontrado: %d\n", freeInodeNum)

	// ? Encontrar bloques libres para el contenido
	blocksNeeded := (size + structs.FILE_CONTENT_SIZE - 1) / structs.FILE_CONTENT_SIZE
	if blocksNeeded == 0 {
		blocksNeeded = 1 // Al menos un bloque para archivos vacíos
	}
	fmt.Printf("📁 DEBUG: Bloques necesarios: %d (tamaño: %d, FILE_CONTENT_SIZE: %d)\n", blocksNeeded, size, structs.FILE_CONTENT_SIZE)

	freeBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, blocksNeeded)
	if err != nil {
		return err
	}
	fmt.Printf("📁 DEBUG: Bloques libres encontrados: %v\n", freeBlocks)

	// ? Crear el inodo del archivo
	now := time.Now()
	timeBytes := FormatTimeToBytes(now)

	fileInode := structs.Inode{
		I_uid:   user.UID,
		I_gid:   user.GID,
		I_size:  size,
		I_atime: timeBytes,
		I_ctime: timeBytes,
		I_mtime: timeBytes,
		I_type:  0, // 0 = archivo
		I_perm:  664,
	}

	// ! Asignar bloques al inodo (con soporte para bloques indirectos)
	err = assignBlocksToInode(file, &fileInode, freeBlocks, blockPos, bmBlockPos, blockCount)
	if err != nil {
		return err
	} // ? Escribir el inodo
	var inodeBuffer bytes.Buffer
	binary.Write(&inodeBuffer, binary.LittleEndian, fileInode)
	_, err = file.WriteAt(inodeBuffer.Bytes(), int64(inodePos+(freeInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return fmt.Errorf("error al escribir el inodo del archivo: %v", err)
	}

	// ? Escribir el contenido en los bloques usando la estructura del inodo
	err = WriteContentToInodeBlocks(file, content, &fileInode, blockPos)
	if err != nil {
		return err
	}

	// ? Marcar el inodo como usado en el bitmap
	err = MarkInodeAsUsed(file, bmInodePos, freeInodeNum)
	if err != nil {
		return err
	}

	// ? Agregar la entrada al directorio padre
	err = AddFileToDirectory(file, fileName, freeInodeNum, parentInodeNum, inodePos, blockPos, bmBlockPos, blockCount)
	if err != nil {
		return err
	}

	return nil
}

func FormatTimeToBytes(t time.Time) [16]byte {
	var timeBytes [16]byte
	timeStr := t.Format("2006-01-02 15:04")
	copy(timeBytes[:], timeStr)
	return timeBytes
}

// assignBlocksToInode asigna bloques al inodo con soporte para bloques indirectos
func assignBlocksToInode(file *os.File, inode *structs.Inode, blocks []int32, blockPos, bmBlockPos, blockCount int32) error {
	fmt.Printf("🔧 DEBUG assignBlocksToInode: Total bloques a asignar: %d\n", len(blocks))
	blockIndex := 0

	// 1. Asignar bloques directos (slots 0-11)
	fmt.Printf("🔧 DEBUG: Asignando bloques directos (slots 0-11)\n")
	for i := 0; i < 12 && blockIndex < len(blocks); i++ {
		inode.I_block[i] = blocks[blockIndex]
		fmt.Printf("🔧 DEBUG: Slot directo %d -> Bloque %d\n", i, blocks[blockIndex])
		blockIndex++
	}

	fmt.Printf("🔧 DEBUG: Bloques directos asignados: %d, restantes: %d\n", blockIndex, len(blocks)-blockIndex)

	// Si se asignaron todos los bloques, terminar
	if blockIndex >= len(blocks) {
		fmt.Printf("🔧 DEBUG: Todos los bloques asignados con bloques directos\n")
		return nil
	}

	// 2. Bloque indirecto simple (slot 12)
	if blockIndex < len(blocks) {
		fmt.Printf("🔧 DEBUG: Necesario bloque indirecto simple (slot 12)\n")
		// Necesitamos un bloque para almacenar punteros
		ptrBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, 1)
		if err != nil {
			return fmt.Errorf("error al encontrar bloque para punteros indirectos: %v", err)
		}

		ptrBlockNum := ptrBlocks[0]
		inode.I_block[12] = ptrBlockNum
		fmt.Printf("🔧 DEBUG: Bloque de punteros indirectos asignado: %d\n", ptrBlockNum)

		// Marcar el bloque de punteros como usado
		err = MarkBlockAsUsed(file, bmBlockPos, ptrBlockNum)
		if err != nil {
			return err
		}

		// Crear bloque de punteros
		var pointerBlock structs.PointerBlock

		// Asignar hasta 16 bloques indirectos
		ptrIndex := 0
		fmt.Printf("🔧 DEBUG: Asignando bloques indirectos:\n")
		for ptrIndex < 16 && blockIndex < len(blocks) {
			pointerBlock.B_pointers[ptrIndex] = blocks[blockIndex]
			fmt.Printf("🔧 DEBUG: Puntero indirecto %d -> Bloque %d\n", ptrIndex, blocks[blockIndex])
			ptrIndex++
			blockIndex++
		}

		fmt.Printf("🔧 DEBUG: Total punteros indirectos asignados: %d\n", ptrIndex)

		// Debug: Verificar el contenido del bloque de punteros antes de escribir
		fmt.Printf("🔧 DEBUG: Contenido del bloque de punteros a escribir: %v\n", pointerBlock.B_pointers)

		// Escribir el bloque de punteros
		var ptrBuffer bytes.Buffer
		binary.Write(&ptrBuffer, binary.LittleEndian, pointerBlock)
		writePosition := int64(blockPos + (ptrBlockNum * structs.BLOCK_SIZE))

		fmt.Printf("🔧 DEBUG: Escribiendo %d bytes en posición %d\n", len(ptrBuffer.Bytes()), writePosition)

		_, err = file.WriteAt(ptrBuffer.Bytes(), writePosition)
		if err != nil {
			return fmt.Errorf("error al escribir bloque de punteros indirectos: %v", err)
		}
		fmt.Printf("🔧 DEBUG: Bloque de punteros escrito exitosamente\n")

		// Debug: Verificar inmediatamente lo que se escribió
		verifyBytes := make([]byte, structs.BLOCK_SIZE)
		_, err = file.ReadAt(verifyBytes, writePosition)
		if err != nil {
			fmt.Printf("🔧 DEBUG ERROR: No se pudo verificar la escritura: %v\n", err)
		} else {
			var verifyBlock structs.PointerBlock
			verifyBuffer := bytes.NewReader(verifyBytes)
			binary.Read(verifyBuffer, binary.LittleEndian, &verifyBlock)
			fmt.Printf("🔧 DEBUG: Verificación - Punteros leídos después de escribir: %v\n", verifyBlock.B_pointers)
		}
	}

	// Si se asignaron todos los bloques, terminar
	if blockIndex >= len(blocks) {
		fmt.Printf("🔧 DEBUG: Todos los bloques asignados (directos + indirectos simples)\n")
		return nil
	}

	// 3. Bloque doblemente indirecto (slot 13)
	if blockIndex < len(blocks) {
		// Necesitamos un bloque para el primer nivel de punteros
		firstLevelBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, 1)
		if err != nil {
			return fmt.Errorf("error al encontrar bloque para punteros doblemente indirectos: %v", err)
		}

		firstLevelBlockNum := firstLevelBlocks[0]
		inode.I_block[13] = firstLevelBlockNum

		err = MarkBlockAsUsed(file, bmBlockPos, firstLevelBlockNum)
		if err != nil {
			return err
		}

		var firstLevelPointers structs.PointerBlock
		firstLevelIndex := 0

		// Para cada entrada en el primer nivel
		for firstLevelIndex < 16 && blockIndex < len(blocks) {
			// Necesitamos un bloque para el segundo nivel
			secondLevelBlocks, err := FindFreeBlocks(file, bmBlockPos, blockCount, 1)
			if err != nil {
				return fmt.Errorf("error al encontrar bloque para segundo nivel: %v", err)
			}

			secondLevelBlockNum := secondLevelBlocks[0]
			firstLevelPointers.B_pointers[firstLevelIndex] = secondLevelBlockNum

			err = MarkBlockAsUsed(file, bmBlockPos, secondLevelBlockNum)
			if err != nil {
				return err
			}

			// Llenar el bloque de segundo nivel
			var secondLevelPointers structs.PointerBlock
			secondLevelIndex := 0

			for secondLevelIndex < 16 && blockIndex < len(blocks) {
				secondLevelPointers.B_pointers[secondLevelIndex] = blocks[blockIndex]
				secondLevelIndex++
				blockIndex++
			}

			// Escribir el bloque de segundo nivel
			var secondBuffer bytes.Buffer
			binary.Write(&secondBuffer, binary.LittleEndian, secondLevelPointers)
			_, err = file.WriteAt(secondBuffer.Bytes(), int64(blockPos+(secondLevelBlockNum*structs.BLOCK_SIZE)))
			if err != nil {
				return fmt.Errorf("error al escribir bloque de segundo nivel: %v", err)
			}

			firstLevelIndex++
		}

		// Escribir el bloque de primer nivel
		var firstBuffer bytes.Buffer
		binary.Write(&firstBuffer, binary.LittleEndian, firstLevelPointers)
		_, err = file.WriteAt(firstBuffer.Bytes(), int64(blockPos+(firstLevelBlockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return fmt.Errorf("error al escribir bloque de primer nivel: %v", err)
		}
	}

	// Si aún quedan bloques, necesitaríamos implementar triple indirecto (slot 14)
	// Por ahora, consideramos que con doble indirecto es suficiente
	if blockIndex < len(blocks) {
		return fmt.Errorf("archivo demasiado grande: se necesitan %d bloques, pero solo se soportan hasta %d", len(blocks), blockIndex)
	}

	return nil
}
