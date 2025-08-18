package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"main/structs"
	"os"
	"strings"
)

const (
	PRIMARY  byte = 'P'
	EXTENDED byte = 'E'
	LOGICAL  byte = 'L'
)
const (
	BEST_FIT  = "BF"
	FIRST_FIT = "FF"
	WORST_FIT = "WF"
)

func CreatePartition(params *structs.FdiskParams) error {
	diskFile, err := os.OpenFile(params.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer diskFile.Close()

	const MBR_SIZE = 47

	mbrBytes := make([]byte, MBR_SIZE)

	_, err = diskFile.ReadAt(mbrBytes, 0)
	if err != nil {
		return fmt.Errorf("error al leer el MBR: %v", err)
	}

	var mbr structs.MBR
	buffer := bytes.NewReader(mbrBytes)

	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_size)
	buffer.Read(mbr.Mbr_timestamp[:])
	buffer.Read(mbr.Mbr_uuid[:])
	binary.Read(buffer, binary.LittleEndian, &mbr.Mbr_partitions)
	buffer.Read(mbr.Dsk_fit[:])

	partitionDataOffset := int64(MBR_SIZE)
	partitionSize := int64(59) // Tamaño de cada estructura PARTITION

	for i := 0; i < 4; i++ {
		partPos := partitionDataOffset + int64(i)*partitionSize

		partData := make([]byte, partitionSize)
		_, err = diskFile.ReadAt(partData, partPos)
		if err != nil {
			return fmt.Errorf("error al leer la partición %d: %v", i, err)
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

		if mbr.Partitions[i].Part_status[0] == 1 {
			partName := string(bytes.Trim(mbr.Partitions[i].Part_name[:], "\x00"))
			if strings.EqualFold(partName, params.Name) {
				return fmt.Errorf("ya existe una partición con el nombre '%s'", params.Name)
			}
		}
	}

	if params.Type == EXTENDED {
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Part_status[0] == 1 && mbr.Partitions[i].Part_type[0] == EXTENDED {
				return errors.New("ya existe una partición extendida en el disco")
			}
		}
	} else if params.Type == LOGICAL {
		extendedExists := false
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Part_status[0] == 1 && mbr.Partitions[i].Part_type[0] == EXTENDED {
				extendedExists = true
				break
			}
		}
		if !extendedExists {
			return errors.New("no se puede crear una partición lógica sin una partición extendida")
		}
	}

	if params.Type != LOGICAL && mbr.Mbr_partitions >= 4 {
		return errors.New("ya se alcanzó el número máximo de particiones primarias y extendidas (4)")
	}

	sizeInBytes := params.GetSizeInBytes()

	if params.Type == LOGICAL {
		return fmt.Errorf("la implementación de particiones lógicas aún no está completa")
	} else {
		var newPartIndex int = -1
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Part_status[0] == 0 {
				newPartIndex = i
				break
			}
		}

		if newPartIndex == -1 {
			return errors.New("no hay slots libres para crear una nueva partición")
		}

		mbr.Partitions[newPartIndex].Part_status[0] = 1
		mbr.Partitions[newPartIndex].Part_type[0] = params.Type
		mbr.Partitions[newPartIndex].Part_size = sizeInBytes

		switch params.Fit {
		case BEST_FIT:
			mbr.Partitions[newPartIndex].Part_fit[0] = 'B'
		case WORST_FIT:
			mbr.Partitions[newPartIndex].Part_fit[0] = 'W'
		default:
			mbr.Partitions[newPartIndex].Part_fit[0] = 'F'
		}

		copy(mbr.Partitions[newPartIndex].Part_name[:], params.Name)

		uuidBytes := GenerateUUID()
		copy(mbr.Partitions[newPartIndex].Part_uuid[:], uuidBytes[:16])

		mbr.Partitions[newPartIndex].Part_correlative = mbr.Mbr_partitions + 1
		mbr.Mbr_partitions++

		mbr.Partitions[newPartIndex].Part_start = int32(MBR_SIZE) + int32(4*int(partitionSize))
		var mbrBuffer bytes.Buffer
		binary.Write(&mbrBuffer, binary.LittleEndian, mbr.Mbr_size)
		mbrBuffer.Write(mbr.Mbr_timestamp[:])
		mbrBuffer.Write(mbr.Mbr_uuid[:])
		binary.Write(&mbrBuffer, binary.LittleEndian, mbr.Mbr_partitions)
		mbrBuffer.Write(mbr.Dsk_fit[:])

		_, err = diskFile.WriteAt(mbrBuffer.Bytes(), 0)
		if err != nil {
			return fmt.Errorf("error al actualizar el MBR: %v", err)
		}

		partPos := partitionDataOffset + int64(newPartIndex)*partitionSize
		var partBuffer bytes.Buffer

		partBuffer.Write(mbr.Partitions[newPartIndex].Part_status[:])
		partBuffer.Write(mbr.Partitions[newPartIndex].Part_type[:])
		partBuffer.Write(mbr.Partitions[newPartIndex].Part_fit[:])
		binary.Write(&partBuffer, binary.LittleEndian, mbr.Partitions[newPartIndex].Part_start)
		binary.Write(&partBuffer, binary.LittleEndian, mbr.Partitions[newPartIndex].Part_size)
		partBuffer.Write(mbr.Partitions[newPartIndex].Part_name[:])
		partBuffer.Write(mbr.Partitions[newPartIndex].Part_uuid[:])
		binary.Write(&partBuffer, binary.LittleEndian, mbr.Partitions[newPartIndex].Part_correlative)

		_, err = diskFile.WriteAt(partBuffer.Bytes(), partPos)
		if err != nil {
			return fmt.Errorf("error al escribir la partición: %v", err)
		}
	}

	fmt.Printf("Partición '%s' creada exitosamente en el disco '%s'\n", params.Name, params.Path)
	return nil
}
