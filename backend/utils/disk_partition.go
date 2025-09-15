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

// calculateAvailableSpace calcula el espacio disponible en el disco
func calculateAvailableSpace(mbr *structs.MBR) int32 {
	diskSize := mbr.Mbr_size
	const MBR_SIZE = 47 + (4 * 59) // MBR + espacio para 4 particiones

	usedSpace := int32(MBR_SIZE)

	// Sumar el espacio usado por particiones primarias y extendidas
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Part_status[0] == 1 {
			usedSpace += mbr.Partitions[i].Part_size
		}
	}

	return diskSize - usedSpace
}

// calculateAvailableSpaceInExtended calcula el espacio disponible en una partición extendida
func calculateAvailableSpaceInExtended(diskFile *os.File, extendedPartition *structs.PARTITION) (int32, error) {
	if extendedPartition == nil {
		return 0, errors.New("partición extendida no válida")
	}

	extendedStart := extendedPartition.Part_start
	extendedSize := extendedPartition.Part_size
	usedSpace := int32(0)

	// Recorrer todas las particiones lógicas para calcular el espacio usado
	currentPos := extendedStart
	ebrSize := int32(30)

	for {
		// Leer el EBR en la posición actual
		ebrData := make([]byte, ebrSize)
		_, err := diskFile.ReadAt(ebrData, int64(currentPos))
		if err != nil {
			break
		}

		// Verificar si hay una partición lógica en esta posición
		partSize := int32(ebrData[6]) | int32(ebrData[7])<<8 | int32(ebrData[8])<<16 | int32(ebrData[9])<<24
		nextEBR := int32(ebrData[10]) | int32(ebrData[11])<<8 | int32(ebrData[12])<<16 | int32(ebrData[13])<<24

		if partSize == 0 {
			break
		}

		// Sumar el espacio usado (EBR + partición)
		usedSpace += ebrSize + partSize

		// Si no hay siguiente EBR, terminar
		if nextEBR == 0 {
			break
		}

		// Ir al siguiente EBR
		currentPos = extendedStart + nextEBR
	}

	return extendedSize - usedSpace, nil
}

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

	// Calcular el tamaño solicitado en bytes
	sizeInBytes := params.GetSizeInBytes()

	// Verificar espacio disponible según el tipo de partición
	if params.Type == LOGICAL {
		// Para particiones lógicas, verificar espacio en la partición extendida
		var extendedPartition *structs.PARTITION
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Part_status[0] == 1 && mbr.Partitions[i].Part_type[0] == EXTENDED {
				extendedPartition = &mbr.Partitions[i]
				break
			}
		}

		if extendedPartition == nil {
			return errors.New("no se puede crear una partición lógica sin una partición extendida")
		}

		availableSpace, err := calculateAvailableSpaceInExtended(diskFile, extendedPartition)
		if err != nil {
			return fmt.Errorf("error al calcular espacio disponible en partición extendida: %v", err)
		}

		// Necesitamos espacio para el EBR (30 bytes) + la partición
		requiredSpace := int32(30) + sizeInBytes
		if availableSpace < requiredSpace {
			return fmt.Errorf("no hay espacio suficiente en la partición extendida. Disponible: %d bytes, Requerido: %d bytes", availableSpace, requiredSpace)
		}
	} else {
		// Para particiones primarias y extendidas, verificar espacio total del disco
		availableSpace := calculateAvailableSpace(&mbr)
		if availableSpace < sizeInBytes {
			return fmt.Errorf("no hay espacio suficiente en el disco. Disponible: %d bytes, Requerido: %d bytes", availableSpace, sizeInBytes)
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

	if params.Type == LOGICAL {
		// Implementación de particiones lógicas
		return createLogicalPartition(diskFile, &mbr, params, sizeInBytes)
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

		// Buscar la siguiente posición disponible después de las particiones existentes
		nextAvailableStart := int32(MBR_SIZE)
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Part_status[0] != 0 && mbr.Partitions[i].Part_size > 0 {
				partEnd := mbr.Partitions[i].Part_start + mbr.Partitions[i].Part_size
				if partEnd > nextAvailableStart {
					nextAvailableStart = partEnd
				}
			}
		}

		mbr.Partitions[newPartIndex].Part_start = nextAvailableStart
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

		// CRÍTICO: Si es una partición extendida, crear el primer EBR vacío
		if params.Type == EXTENDED {
			err = createInitialEBR(diskFile, mbr.Partitions[newPartIndex].Part_start)
			if err != nil {
				return fmt.Errorf("error al crear el EBR inicial: %v", err)
			}
		}
	}

	fmt.Printf("Partición '%s' creada exitosamente en el disco '%s'\n", params.Name, params.Path)
	return nil
}

// createLogicalPartition crea una partición lógica dentro de una partición extendida
func createLogicalPartition(diskFile *os.File, mbr *structs.MBR, params *structs.FdiskParams, sizeInBytes int32) error {
	// Buscar la partición extendida
	var extendedPartition *structs.PARTITION
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Part_status[0] == 1 && mbr.Partitions[i].Part_type[0] == EXTENDED {
			extendedPartition = &mbr.Partitions[i]
			break
		}
	}

	if extendedPartition == nil {
		return errors.New("no se encontró una partición extendida para crear la partición lógica")
	}

	// Verificar que no existe una partición con el mismo nombre (primarias/extendidas)
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Part_status[0] == 1 {
			partName := string(bytes.Trim(mbr.Partitions[i].Part_name[:], "\x00"))
			if strings.EqualFold(partName, params.Name) {
				return fmt.Errorf("ya existe una partición con el nombre '%s'", params.Name)
			}
		}
	}

	// Calcular el espacio disponible en la partición extendida
	extendedStart := extendedPartition.Part_start

	// Buscar la posición donde crear la nueva partición lógica
	currentPos := extendedStart
	ebrSize := int32(30)    // Tamaño del EBR
	lastEBRPos := int32(-1) // Posición del último EBR para actualizar su Part_next

	// Leer todas las particiones lógicas existentes para encontrar el último EBR
	for {
		// Leer el EBR en la posición actual
		ebrData := make([]byte, ebrSize)
		_, err := diskFile.ReadAt(ebrData, int64(currentPos))
		if err != nil {
			break
		}

		// Verificar si hay una partición lógica en esta posición
		partSize := int32(ebrData[6]) | int32(ebrData[7])<<8 | int32(ebrData[8])<<16 | int32(ebrData[9])<<24
		nextEBR := int32(ebrData[10]) | int32(ebrData[11])<<8 | int32(ebrData[12])<<16 | int32(ebrData[13])<<24

		// Si no hay partición aquí, podemos usar este espacio
		if partSize == 0 {
			break
		}

		// Verificar duplicado de nombre
		partName := string(bytes.Trim(ebrData[14:30], "\x00"))
		if strings.EqualFold(partName, params.Name) {
			return fmt.Errorf("ya existe una partición lógica con el nombre '%s'", params.Name)
		}

		// Si no hay siguiente EBR, podemos crear después de esta partición
		if nextEBR == 0 {
			lastEBRPos = currentPos // Guardar la posición del último EBR
			currentPos = currentPos + ebrSize + partSize
			break
		}

		// Ir al siguiente EBR
		currentPos = extendedStart + nextEBR
	}

	// Crear el nuevo EBR
	var ebr structs.EBR
	ebr.Part_mount[0] = 0 // No montado

	switch params.Fit {
	case BEST_FIT:
		ebr.Part_fit[0] = 'B'
	case WORST_FIT:
		ebr.Part_fit[0] = 'W'
	default:
		ebr.Part_fit[0] = 'F'
	}

	ebr.Part_start = ebrSize // La partición empieza después del EBR
	ebr.Part_size = sizeInBytes
	ebr.Part_next = 0 // Por ahora no hay siguiente (se actualizará si se crea otra)
	copy(ebr.Part_name[:], params.Name)

	// Escribir el EBR al disco
	var ebrBuffer bytes.Buffer
	ebrBuffer.Write(ebr.Part_mount[:])
	ebrBuffer.Write(ebr.Part_fit[:])
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_start)
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_size)
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_next)
	ebrBuffer.Write(ebr.Part_name[:])

	_, err := diskFile.WriteAt(ebrBuffer.Bytes(), int64(currentPos))
	if err != nil {
		return fmt.Errorf("error al escribir el EBR: %v", err)
	}

	// CRÍTICO: Actualizar el EBR anterior para que apunte al nuevo EBR
	if lastEBRPos != -1 {
		// Calcular la posición relativa del nuevo EBR desde el inicio de la partición extendida
		newEBRRelativePos := currentPos - extendedStart

		// Leer el EBR anterior
		lastEBRData := make([]byte, ebrSize)
		_, err := diskFile.ReadAt(lastEBRData, int64(lastEBRPos))
		if err != nil {
			return fmt.Errorf("error al leer el EBR anterior: %v", err)
		}

		// Actualizar el campo Part_next del EBR anterior
		var lastEBRBuffer bytes.Buffer
		lastEBRBuffer.Write(lastEBRData[:10])                                // Part_mount, Part_fit, Part_start, Part_size (10 bytes)
		binary.Write(&lastEBRBuffer, binary.LittleEndian, newEBRRelativePos) // Part_next (4 bytes)
		lastEBRBuffer.Write(lastEBRData[14:])                                // Part_name (16 bytes)

		// Escribir el EBR anterior actualizado
		_, err = diskFile.WriteAt(lastEBRBuffer.Bytes(), int64(lastEBRPos))
		if err != nil {
			return fmt.Errorf("error al actualizar el EBR anterior: %v", err)
		}
	}

	// No incrementamos mbr.Mbr_partitions para particiones lógicas
	// ya que este contador es solo para particiones primarias y extendidas

	return nil
}

// createInitialEBR crea el primer EBR vacío cuando se crea una partición extendida
func createInitialEBR(diskFile *os.File, extendedStart int32) error {
	// Crear un EBR vacío
	var ebr structs.EBR

	// Inicializar todos los campos en cero/vacío
	ebr.Part_mount[0] = 0 // No montado
	ebr.Part_fit[0] = 0   // Sin ajuste definido
	ebr.Part_start = 0    // Sin partición
	ebr.Part_size = 0     // Sin tamaño
	ebr.Part_next = 0     // Sin siguiente EBR
	// ebr.Part_name ya está inicializado en ceros

	// Escribir el EBR vacío al inicio de la partición extendida
	var ebrBuffer bytes.Buffer
	ebrBuffer.Write(ebr.Part_mount[:])
	ebrBuffer.Write(ebr.Part_fit[:])
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_start)
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_size)
	binary.Write(&ebrBuffer, binary.LittleEndian, ebr.Part_next)
	ebrBuffer.Write(ebr.Part_name[:])

	_, err := diskFile.WriteAt(ebrBuffer.Bytes(), int64(extendedStart))
	if err != nil {
		return fmt.Errorf("error al escribir el EBR inicial: %v", err)
	}

	fmt.Printf("EBR inicial creado en la posición %d\n", extendedStart)
	return nil
}
