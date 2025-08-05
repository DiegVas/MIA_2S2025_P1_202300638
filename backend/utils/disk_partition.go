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

// Las constantes para los tipos de particiones
const (
PRIMARY    byte = 'P'
EXTENDED   byte = 'E'
LOGICAL    byte = 'L'
)

// Las constantes para los tipos de ajuste
const (
BEST_FIT  = "BF"
FIRST_FIT = "FF"
WORST_FIT = "WF"
)

// CreatePartition crea una nueva partición en el disco
func CreatePartition(params *structs.FdiskParams) error {
// Abrir el archivo del disco
diskFile, err := os.OpenFile(params.Path, os.O_RDWR, 0666)
if err != nil {
return fmt.Errorf("error al abrir el disco: %v", err)
}
defer diskFile.Close()

// Definir tamaños fijos para estructuras
const MBR_SIZE = 47 // Tamaño del MBR: int32 + [10]byte + [16]byte + int32 + [1]byte + padding

// Leer el MBR
mbrBytes := make([]byte, MBR_SIZE)

_, err = diskFile.ReadAt(mbrBytes, 0)
if err != nil {
return fmt.Errorf("error al leer el MBR: %v", err)
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
return fmt.Errorf("error al leer la partición %d: %v", i, err)
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

// Verificar si ya existe una partición con el mismo nombre
if mbr.Partitions[i].Part_status[0] == 1 {
partName := string(bytes.Trim(mbr.Partitions[i].Part_name[:], "\x00"))
if strings.ToLower(partName) == strings.ToLower(params.Name) {
return fmt.Errorf("ya existe una partición con el nombre '%s'", params.Name)
}
}
}

// Validar el tipo de partición
if params.Type == EXTENDED {
// Verificar si ya existe una partición extendida
for i := 0; i < 4; i++ {
if mbr.Partitions[i].Part_status[0] == 1 && mbr.Partitions[i].Part_type[0] == EXTENDED {
return errors.New("ya existe una partición extendida en el disco")
}
}
} else if params.Type == LOGICAL {
// Verificar si existe una partición extendida para crear la lógica
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

// Verificar si hay espacio para más particiones primarias/extendidas
if params.Type != LOGICAL && mbr.Mbr_partitions >= 4 {
return errors.New("ya se alcanzó el número máximo de particiones primarias y extendidas (4)")
}

// Calcular el tamaño de la partición en bytes
sizeInBytes := params.GetSizeInBytes()

// Si es una partición lógica, manejarla de forma especial
if params.Type == LOGICAL {
// Buscar la partición extendida y crear la partición lógica dentro de ella
// Esta sección debería implementar la gestión de EBRs
return fmt.Errorf("la implementación de particiones lógicas aún no está completa")
} else {
// Crear una partición primaria o extendida
// Buscar un slot libre
var newPartIndex int = -1
for i := 0; i < 4; i++ {
if mbr.Partitions[i].Part_status[0] == 0 { // Slot libre
newPartIndex = i
break
}
}

if newPartIndex == -1 {
return errors.New("no hay slots libres para crear una nueva partición")
}

// Configurar la nueva partición
mbr.Partitions[newPartIndex].Part_status[0] = 1 // Activa
mbr.Partitions[newPartIndex].Part_type[0] = params.Type
mbr.Partitions[newPartIndex].Part_size = sizeInBytes

// Asignar fit
switch params.Fit {
case BEST_FIT:
mbr.Partitions[newPartIndex].Part_fit[0] = 'B'
case WORST_FIT:
mbr.Partitions[newPartIndex].Part_fit[0] = 'W'
default:
mbr.Partitions[newPartIndex].Part_fit[0] = 'F' // First Fit por defecto
}

// Asignar nombre
copy(mbr.Partitions[newPartIndex].Part_name[:], params.Name)

// Generar y asignar UUID
uuidBytes := generateUUID()
copy(mbr.Partitions[newPartIndex].Part_uuid[:], uuidBytes[:16])

// Asignar correlativo y actualizar contador de particiones
mbr.Partitions[newPartIndex].Part_correlative = mbr.Mbr_partitions + 1
mbr.Mbr_partitions++

        // Calcular la posición inicial de la partición (simplificado)
        // En una implementación completa, debería usar algoritmos de ajuste
        mbr.Partitions[newPartIndex].Part_start = int32(MBR_SIZE) + int32(4 * int(partitionSize))// Escribir el MBR actualizado
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

// Escribir la partición actualizada
partPos := partitionDataOffset + int64(newPartIndex) * partitionSize
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

// Implementación simplificada de findPartitionSpace
func findPartitionSpace(diskFile *os.File, mbr structs.MBR, size int32) (int32, error) {
    // En una implementación real, esto debería usar algoritmos de ajuste
    // y buscar espacios disponibles en el disco

    // Por ahora, devolvemos una posición fija después del MBR y las particiones
    const MBR_SIZE = 47
    const PARTITION_SIZE = 59

    return int32(MBR_SIZE) + int32(4 * PARTITION_SIZE), nil
}