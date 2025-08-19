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

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// ! Crear archivo con el tamaño especificado
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al crear archivo: %v", err)
	}
	defer file.Close()

	// * Obtener el tamaño del disco en bytes
	sizeInBytes := params.GetSizeInBytes()

	// ? Crear MBR
	mbr := structs.MBR{
		Mbr_size:       sizeInBytes,
		Mbr_partitions: 0,
	}

	timestampStr := time.Now().Format("2006-01-02")[:10]
	copy(mbr.Mbr_timestamp[:], timestampStr)

	uuidBytes := GenerateUUID()
	copy(mbr.Mbr_uuid[:], uuidBytes)

	if len(params.Fit) > 0 {
		mbr.Dsk_fit[0] = params.Fit[0]
	} else {
		mbr.Dsk_fit[0] = 'F' // Valor por defecto First Fit
	}

	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, mbr.Mbr_size)

	buf.Write(mbr.Mbr_timestamp[:])

	buf.Write(mbr.Mbr_uuid[:])

	binary.Write(&buf, binary.LittleEndian, mbr.Mbr_partitions)

	for i := 0; i < 4; i++ {

		buf.Write([]byte{0})

		buf.Write([]byte{0})

		buf.Write([]byte{0})

		binary.Write(&buf, binary.LittleEndian, int32(0))
		binary.Write(&buf, binary.LittleEndian, int32(0))
		binary.Write(&buf, binary.LittleEndian, int32(0))

		emptyName := make([]byte, 16)
		buf.Write(emptyName)

		emptyUUID := make([]byte, 16)
		buf.Write(emptyUUID)
	}

	buf.Write(mbr.Dsk_fit[:])

	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error al escribir MBR: %v", err)
	}

	buffer := make([]byte, 1024)

	bytesWritten := int32(buf.Len())
	remainingBytes := sizeInBytes - bytesWritten

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
