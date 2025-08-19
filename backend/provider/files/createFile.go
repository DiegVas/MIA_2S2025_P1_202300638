package files

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"main/provider/mounted"
	"main/structs"
	"main/utils"
)

func CreateFile(params *structs.MkfileParams) error {

	session := structs.GetSession()
	if !session.Active {
		return errors.New("debe iniciar sesión para usar el comando MKFILE")
	}

	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return errors.New("no hay ninguna partición montada para esta sesión")
	}

	mbr, err := utils.ReadMBR(partition.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	// ! Buscar la partición montada
	var targetPartition structs.PARTITION
	var partitionFound bool
	for _, p := range mbr.Partitions {
		partName := strings.TrimRight(string(p.Part_name[:]), "\x00")
		if partName == partition.Name {
			targetPartition = p
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		return fmt.Errorf("la partición '%s' no se encontró en el disco", partition.Name)
	}

	// ? Calcular las posiciones de las estructuras del sistema de archivos
	startPosition := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE

	bmInodePos := startPosition + structs.SUPERBLOCK_SIZE
	bmBlockPos := bmInodePos + inodeCount
	inodePos := bmBlockPos + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)

	// ? Abrir el archivo del disco para escritura
	file, err := os.OpenFile(partition.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	// ? Preparar el contenido del archivo
	content, contentSize, err := prepareFileContent(params)
	if err != nil {
		return err
	}

	// ? Verificar y crear el archivo
	err = createFileInFilesystem(file, params.Path, content, contentSize, session,
		inodePos, blockPos, bmInodePos, bmBlockPos, inodeCount, blockCount, params.R)
	if err != nil {
		return err
	}

	return nil
}
