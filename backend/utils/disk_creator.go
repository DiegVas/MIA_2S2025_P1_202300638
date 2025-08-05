package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"main/structs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Función para generar un UUID aleatorio
func generateUUID() string {
    uuid := make([]byte, 16)
    rand.Read(uuid)
    uuid[8] = uuid[8]&^0xc0 | 0x80
    uuid[6] = uuid[6]&^0xf0 | 0x40
    return hex.EncodeToString(uuid)
}

func CreateDisk(params *structs.MkdiskParams) error {

    // Limpiar comillas de la ruta
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
        Dsk_fit:        params.Fit[0], // Primer carácter del Fit (F, B o W)
    }

    // Asignar timestamp y UUID
    timestampStr := time.Now().Format(time.RFC3339)
    copy(mbr.Mbr_timestamp[:], timestampStr)

    uuidStr := generateUUID()
    copy(mbr.Mbr_uuid[:], uuidStr)

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
    
    // Escribir fit del disco
    binary.Write(&buf, binary.LittleEndian, mbr.Dsk_fit)
    
    // Escribir el MBR completo al inicio del archivo
    _, err = file.Write(buf.Bytes())
    if err != nil {
        return fmt.Errorf("error al escribir MBR: %v", err)
    }
    
    // Crear un buffer de 1024 bytes lleno de ceros para escribir el resto del disco
    buffer := make([]byte, 1024)
    
    // Calcular cuántos bytes quedan por escribir después del MBR
    bytesWritten := int64(buf.Len())
    remainingBytes := sizeInBytes - bytesWritten
    
    // Escribir el resto del disco en bloques de 1024 bytes
    for remainingBytes > 0 {
        writeSize := int64(len(buffer))
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