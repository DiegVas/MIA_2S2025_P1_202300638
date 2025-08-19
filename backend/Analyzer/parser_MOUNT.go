package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
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
