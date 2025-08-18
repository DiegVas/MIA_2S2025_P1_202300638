package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateDisk(params *structs.MkdiskParams) error {

	path := params.Path
	path = strings.Trim(path, "\"'")

	// Asegurarse que el directorio exista
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Crear archivo con el tamaño especificado
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al crear archivo: %v", err)
	}
	defer file.Close()

	// Obtener el tamaño del disco en bytes
	sizeInBytes := params.GetSizeInBytes()

	// Crear MBR
	mbr := structs.MBR{
		Mbr_size:       sizeInBytes,
		Mbr_partitions: 0,
	}

	// Asignar timestamp
	timestampStr := time.Now().Format("2006-01-02")[:10]
	copy(mbr.Mbr_timestamp[:], timestampStr)

	// Asignar UUID
	uuidBytes := GenerateUUID()
	copy(mbr.Mbr_uuid[:], uuidBytes)

	// Asignar fit
	if len(params.Fit) > 0 {
		mbr.Dsk_fit[0] = params.Fit[0]
	} else {
		mbr.Dsk_fit[0] = 'F' // Valor por defecto First Fit
	}

	// Escribir el MBR manualmente campo por campo
	var buf bytes.Buffer

	// Escribir tamaño del disco
	binary.Write(&buf, binary.LittleEndian, mbr.Mbr_size)

	// Escribir timestamp como array de bytes fijo
	buf.Write(mbr.Mbr_timestamp[:])

	// Escribir UUID como array de bytes fijo
	buf.Write(mbr.Mbr_uuid[:])

	// Escribir número de particiones
	binary.Write(&buf, binary.LittleEndian, mbr.Mbr_partitions)

	// Escribir particiones (inicialmente vacías)
	for i := 0; i < 4; i++ {
		// Part_status
		buf.Write([]byte{0}) // Inactiva

		// Part_type
		buf.Write([]byte{0}) // Tipo vacío

		// Part_fit
		buf.Write([]byte{0}) // Fit vacío

		// Part_start, Part_size, Part_correlative
		binary.Write(&buf, binary.LittleEndian, int32(0))
		binary.Write(&buf, binary.LittleEndian, int32(0))
		binary.Write(&buf, binary.LittleEndian, int32(0))

		// Part_name (16 bytes)
		emptyName := make([]byte, 16)
		buf.Write(emptyName)

		// Part_uuid (16 bytes)
		emptyUUID := make([]byte, 16)
		buf.Write(emptyUUID)
	}

	// Escribir fit del disco
	buf.Write(mbr.Dsk_fit[:])

	// Escribir el MBR completo al inicio del archivo
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error al escribir MBR: %v", err)
	}

	// Crear un buffer de 1024 bytes lleno de ceros para escribir el resto del disco
	buffer := make([]byte, 1024)

	// Calcular cuántos bytes quedan por escribir después del MBR
	bytesWritten := int32(buf.Len())
	remainingBytes := sizeInBytes - bytesWritten

	// Escribir el resto del disco en bloques de 1024 bytes
	for remainingBytes > 0 {
		writeSize := int32(len(buffer))
		if remainingBytes < writeSize {
			writeSize = remainingBytes
		}

		_, err = file.Write(buffer[:writeSize])
		if err != nil {
			return fmt.Errorf("error al escribir datos del disco: %v", err)
		}

		remainingBytes -= writeSize
	}

	fmt.Printf("Disco creado exitosamente en: %s (Tamaño: %d bytes)\n", path, sizeInBytes)
	return nil
}
