package mounted

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
)

// ReadMBR lee el Master Boot Record del disco especificado
func ReadMBR(diskPath string) (*structs.MBR, error) {
	diskFile, err := os.OpenFile(diskPath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	const MBR_SIZE = 47

	mbrBytes := make([]byte, MBR_SIZE)

	_, err = diskFile.ReadAt(mbrBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR: %v", err)
	}

	var mbr structs.MBR
	buffer := bytes.NewReader(mbrBytes)

	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_size)
	buffer.Read(mbr.Mbr_timestamp[:])
	buffer.Read(mbr.Mbr_uuid[:])
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_partitions)
	buffer.Read(mbr.Dsk_fit[:])

	partitionDataOffset := int64(MBR_SIZE)
	partitionSize := int64(59)

	for i := 0; i < 4; i++ {
		partPos := partitionDataOffset + int64(i)*partitionSize

		partData := make([]byte, partitionSize)
		_, err = diskFile.ReadAt(partData, partPos)
		if err != nil {
			return nil, fmt.Errorf("error al leer la partición %d: %v", i, err)
		}

		partBuffer := bytes.NewReader(partData)

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

// UpdatePartition actualiza una partición en el disco
func UpdatePartition(diskPath string, partition structs.PARTITION, mbr *structs.MBR) error {
	diskFile, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	const MBR_SIZE = 47

	partitionIndex := -1
	for i := 0; i < 4; i++ {
		partName := string(bytes.Trim(mbr.Partitions[i].Part_name[:], "\x00"))
		updatePartName := string(bytes.Trim(partition.Part_name[:], "\x00"))

		if partName == updatePartName {
			partitionIndex = i
			mbr.Partitions[i] = partition
			break
		}
	}

	if partitionIndex == -1 {
		return fmt.Errorf("no se encontró la partición '%s' en el MBR", string(bytes.Trim(partition.Part_name[:], "\x00")))
	}

	partitionDataOffset := int64(MBR_SIZE)
	partitionSize := int64(59)
	partPos := partitionDataOffset + int64(partitionIndex)*partitionSize

	var partBuffer bytes.Buffer

	partBuffer.Write(partition.Part_status[:])
	partBuffer.Write(partition.Part_type[:])
	partBuffer.Write(partition.Part_fit[:])
	binary.Write(&partBuffer, binary.LittleEndian, partition.Part_start)
	binary.Write(&partBuffer, binary.LittleEndian, partition.Part_size)
	partBuffer.Write(partition.Part_name[:])
	partBuffer.Write(partition.Part_uuid[:])
	binary.Write(&partBuffer, binary.LittleEndian, partition.Part_correlative)

	_, err = diskFile.WriteAt(partBuffer.Bytes(), partPos)
	if err != nil {
		return fmt.Errorf("error al escribir la partición actualizada: %v", err)
	}

	return nil
}
