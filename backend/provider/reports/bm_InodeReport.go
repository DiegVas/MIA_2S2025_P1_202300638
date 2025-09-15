package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"path/filepath"
	"strings"
)

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

func readInodeBitmapFromFile(file *os.File, superblock *structs.Superblock) ([]byte, error) {
	bitmapSize := superblock.S_inodes_count
	bitmap := make([]byte, bitmapSize)

	_, err := file.ReadAt(bitmap, int64(superblock.S_bm_inode_start))
	if err != nil {
		return nil, fmt.Errorf("error leyendo bitmap de inodos: %v", err)
	}

	return bitmap, nil
}

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
