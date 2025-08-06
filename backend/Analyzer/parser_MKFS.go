package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

// ParseMkfs analiza el comando MKFS y extrae los parámetros
func ParseMkfs(command string) (*structs.MkfsParams, error) {
	// Crear la estructura para almacenar los parámetros
	params := &structs.MkfsParams{
		Type: "full", // Valor por defecto
	}
	
	// Preservamos el comando original para mantener mayúsculas/minúsculas en valores
	originalCommand := command
	
	// Convertir a minúsculas solo para verificar el prefijo del comando
	commandLower := strings.ToLower(command)
	
	// Verificar que el comando comience con "mkfs"
	if !strings.HasPrefix(commandLower, "mkfs") {
		return nil, errors.New("el comando debe comenzar con 'mkfs'")
	}

	// Obtener solo los parámetros - usamos el comando original
	paramStr := ""
	if strings.HasPrefix(commandLower, "mkfs ") {
		paramStr = originalCommand[5:] // Remover "mkfs "
	} else {
		paramStr = originalCommand[4:] // Remover "mkfs" sin espacio
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
		"id": false,
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
		case "id":
			params.ID = paramValue // Mantener mayúsculas/minúsculas originales
			requiredParams["id"] = true
			
		case "type":
			typeValue := strings.ToLower(paramValue)
			if typeValue != "full" {
				return nil, fmt.Errorf("tipo de formateo inválido: '%s'. Solo se admite 'full'", paramValue)
			}
			params.Type = typeValue
		
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
