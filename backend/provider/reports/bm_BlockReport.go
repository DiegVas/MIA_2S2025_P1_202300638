package reports

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"os"
	"path/filepath"
	"strings"
)

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

func readBlockBitmapFromFile(file *os.File, superblock *structs.Superblock) ([]byte, error) {
	bitmapSize := superblock.S_blocks_count
	bitmap := make([]byte, bitmapSize)

	_, err := file.ReadAt(bitmap, int64(superblock.S_bm_block_start))
	if err != nil {
		return nil, fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	return bitmap, nil
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
