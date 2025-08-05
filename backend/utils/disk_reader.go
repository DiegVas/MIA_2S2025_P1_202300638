package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
)

// ReadMBR lee el MBR de un disco y devuelve la estructura
func ReadMBR(diskPath string) (*structs.MBR, error) {
	// Abrir el archivo del disco
	diskFile, err := os.OpenFile(diskPath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	// Definir tamaños fijos para estructuras
	const MBR_SIZE = 47 // Tamaño del MBR: int32 + [10]byte + [16]byte + int32 + [1]byte + padding

	// Leer el MBR
	mbrBytes := make([]byte, MBR_SIZE)

	_, err = diskFile.ReadAt(mbrBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Deserializar el MBR
	var mbr structs.MBR
	buffer := bytes.NewReader(mbrBytes)

	// Leer cada campo individualmente
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_size)
	buffer.Read(mbr.Mbr_timestamp[:])
	buffer.Read(mbr.Mbr_uuid[:])
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_partitions)
	buffer.Read(mbr.Dsk_fit[:])

	// Leer las particiones (ahora están en un array dentro del MBR)
	partitionDataOffset := int64(MBR_SIZE)
	partitionSize := int64(59) // Tamaño de cada estructura PARTITION

	for i := 0; i < 4; i++ {
		// Calcular la posición de cada partición
		partPos := partitionDataOffset + int64(i) * partitionSize

		// Leer los datos de la partición
		partData := make([]byte, partitionSize)
		_, err = diskFile.ReadAt(partData, partPos)
		if err != nil {
			return nil, fmt.Errorf("error al leer la partición %d: %v", i, err)
		}

		partBuffer := bytes.NewReader(partData)

		// Leer los campos de la partición
		partBuffer.Read(mbr.Partitions[i].Part_status[:])
		partBuffer.Read(mbr.Partitions[i].Part_type[:])
		partBuffer.Read(mbr.Partitions[i].Part_fit[:])
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_start)
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_size)
		partBuffer.Read(mbr.Partitions[i].Part_name[:])
		partBuffer.Read(mbr.Partitions[i].Part_uuid[:])
		binary.Read(partBuffer, binary.LittleEndian, &mbr.Partitions[i].Part_correlative)
	}

	return &mbr, nil
}
