package mounted

import (
	"fmt"
	"main/structs"
	"os"
	"strings"
)

func MountPartition(params *structs.MountParams) (string, error) {
	// Verificar que el disco exista
	if _, err := os.Stat(params.Path); os.IsNotExist(err) {
		return "", fmt.Errorf("el disco en la ruta '%s' no existe", params.Path)
	}

	// Obtener el gestor de montajes
	mountManager := GetMountManager()

	// Verificar si la partición ya está montada
	if mountManager.IsMounted(params.Path, params.Name) {
		return "", fmt.Errorf("la partición '%s' del disco '%s' ya está montada", params.Name, params.Path)
	}

	// Verificar que la partición exista en el disco
	mbr, err := ReadMBR(params.Path)
	if err != nil {
		return "", fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	// Buscar la partición por nombre
	partitionExists := false
	var targetPartition structs.PARTITION

	for _, partition := range mbr.Partitions {
		// Convertir el nombre de la partición a string (eliminando caracteres nulos)
		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")

		// Comparación sensible a mayúsculas/minúsculas
		if partName == params.Name {
			// Verificar que la partición sea primaria (o lógica, pero según los requisitos solo se usarán primarias)
			partType := string(partition.Part_type[:])
			if partType != "P" && partType != "p" {
				return "", fmt.Errorf("solo se pueden montar particiones primarias, '%s' es de tipo %s", params.Name, partType)
			}

			// Ya no verificamos si está activa, simplemente la marcaremos como activa al montarla
			partitionExists = true
			targetPartition = partition
			break
		}
	}

	if !partitionExists {
		return "", fmt.Errorf("la partición '%s' no existe en el disco '%s'", params.Name, params.Path)
	}

	// Usar los últimos dos dígitos del carnet fijo 202300638
	carnet := "38"

	// Montar la partición y obtener el ID
	id, correlative, err := mountManager.AddMountedPartition(params.Path, params.Name, carnet)
	if err != nil {
		return "", fmt.Errorf("error al montar la partición: %v", err)
	}

	// Actualizar el correlativo en la partición
	targetPartition.Part_correlative = int32(correlative)

	// Actualizar el status a montado (A)
	targetPartition.Part_status[0] = 'A'

	// Actualizar la partición en el disco
	err = UpdatePartition(params.Path, targetPartition, mbr)
	if err != nil {
		return "", fmt.Errorf("error al actualizar la partición en el disco: %v", err)
	}

	// Guardar el ID en los parámetros
	params.ID = id

	return id, nil
}
