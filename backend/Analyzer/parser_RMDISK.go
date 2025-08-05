package Analyzer

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type RmdiskParams struct {
	Path string
}

func ParseRmdisk(command string) (*RmdiskParams, error) {
	// Inicializar parámetros
	params := &RmdiskParams{
		Path: "",
	}

	command = strings.TrimSpace(command)
	if !strings.HasPrefix(strings.ToLower(command), "rmdisk") {
		return nil, errors.New("comando inválido: debe comenzar con 'rmdisk'")
	}

	// Obtener solo los parámetros
	paramStr := strings.TrimSpace(command[6:]) // Remover "rmdisk"

	// Expresión regular mejorada para capturar parámetros con valores que pueden contener espacios
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	pathFound := false

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		paramName := strings.ToLower(match[1])

		// Determinar el valor del parámetro dependiendo de qué grupo de captura contiene el valor
		var paramValue string
		if match[2] != "" {
			paramValue = match[2] // Valor con comillas dobles
		} else if match[3] != "" {
			paramValue = match[3] // Valor con comillas simples
		} else {
			paramValue = match[4] // Valor sin comillas
		}

		switch paramName {
		case "path":
			if paramValue == "" {
				return nil, errors.New("el parámetro path no puede estar vacío")
			}
			params.Path = paramValue
			pathFound = true

		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", paramName)
		}
	}

	if !pathFound {
		return nil, errors.New("el parámetro -path es obligatorio")
	}

	return params, nil
}

func RemoveDisk(params *RmdiskParams) error {
	// Limpiar comillas de la ruta
	path := params.Path
	path = strings.Trim(path, "\"'")

	// Verificar si el archivo existe
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el disco no existe en la ruta especificada: %s", path)
		}
		return fmt.Errorf("error al verificar si el disco existe: %v", err)
	}

	// Eliminar el archivo
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("error al eliminar el disco: %v", err)
	}

	fmt.Printf("Disco eliminado exitosamente: %s\n", path)
	return nil
}