package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"main/utils"
	"os"
	"regexp"
	"strings"
)

// ParseMount analiza el comando MOUNT y extrae los parámetros
func ParseMount(command string) (*structs.MountParams, error) {
	// Crear la estructura para almacenar los parámetros
	params := &structs.MountParams{
		// Valor fijo para el carnet
		Carnet: "202300638",
	}
	
	// Preservamos el comando original para mantener mayúsculas/minúsculas en valores
	originalCommand := command
	
	// Convertir a minúsculas solo para verificar el prefijo del comando
	commandLower := strings.ToLower(command)
	
	// Verificar que el comando comience con "mount"
	if !strings.HasPrefix(commandLower, "mount") {
		return nil, errors.New("el comando debe comenzar con 'mount'")
	}

	// Obtener solo los parámetros - usamos el comando original
	paramStr := ""
	if strings.HasPrefix(commandLower, "mount ") {
		paramStr = originalCommand[6:] // Remover "mount "
	} else {
		paramStr = originalCommand[5:] // Remover "mount" sin espacio
	}
	paramStr = strings.TrimSpace(paramStr)

	// Expresión regular para capturar parámetros con valores que pueden contener espacios
	// Captura: -parámetro="valor con espacios" o -parámetro='valor con espacios' o -parámetro=valor_sin_espacios
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	
	// Mapa para verificar parámetros obligatorios
	requiredParams := map[string]bool{
		"path": false,
		"name": false,
	}
	
	// Procesar los parámetros encontrados
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		// El nombre del parámetro se convierte a minúsculas para ser insensible a mayúsculas/minúsculas
		paramName := strings.ToLower(match[1])
		
		// Determinar el valor del parámetro dependiendo de qué grupo de captura contiene el valor
		// NO convertimos el valor a minúsculas para mantener sensibilidad a mayúsculas/minúsculas
		var paramValue string
		if match[2] != "" {
			paramValue = match[2] // Valor con comillas dobles
		} else if match[3] != "" {
			paramValue = match[3] // Valor con comillas simples
		} else if match[4] != "" {
			paramValue = match[4] // Valor sin comillas
		}
		
		switch paramName {
		case "path":
			params.Path = paramValue // Mantener mayúsculas/minúsculas originales
			requiredParams["path"] = true
			
			// Verificar que el disco exista
			if _, err := os.Stat(paramValue); os.IsNotExist(err) {
				return nil, fmt.Errorf("el disco en la ruta '%s' no existe", paramValue)
			}
			
		case "name":
			params.Name = paramValue // Mantener mayúsculas/minúsculas originales
			requiredParams["name"] = true
		
		// Eliminamos el caso "carnet" ya que ahora usamos un valor fijo
		
		default:
			return nil, fmt.Errorf("parámetro desconocido: '%s'", paramName)
		}
	}
	
	// Verificar que todos los parámetros obligatorios estén presentes
	for param, found := range requiredParams {
		if !found {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", param)
		}
	}
	
	return params, nil
}

// MountPartition monta una partición y devuelve el ID asignado
func MountPartition(params *structs.MountParams) (string, error) {
	// Verificar que el disco exista
	if _, err := os.Stat(params.Path); os.IsNotExist(err) {
		return "", fmt.Errorf("el disco en la ruta '%s' no existe", params.Path)
	}

	// Obtener el gestor de montajes
	mountManager := structs.GetMountManager()
	
	// Verificar si la partición ya está montada
	if mountManager.IsMounted(params.Path, params.Name) {
		return "", fmt.Errorf("la partición '%s' del disco '%s' ya está montada", params.Name, params.Path)
	}
	
	// Verificar que la partición exista en el disco
	mbr, err := utils.ReadMBR(params.Path)
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
	err = utils.UpdatePartition(params.Path, targetPartition, mbr)
	if err != nil {
		return "", fmt.Errorf("error al actualizar la partición en el disco: %v", err)
	}
	
	// Guardar el ID en los parámetros
	params.ID = id
	
	return id, nil
}
