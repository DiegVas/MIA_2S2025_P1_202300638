package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"os"
	"regexp"
	"strings"
)

func ParseRmdisk(command string) (*structs.RmdiskParams, error) {
	params := &structs.RmdiskParams{
		Path: "",
	}

	command = strings.TrimSpace(command)
	if !strings.HasPrefix(strings.ToLower(command), "rmdisk") {
		return nil, errors.New("comando inválido: debe comenzar con 'rmdisk'")
	}

	paramStr := strings.TrimSpace(command[6:])

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

		var paramValue string
		if match[2] != "" {
			paramValue = match[2]
		} else if match[3] != "" {
			paramValue = match[3]
		} else {
			paramValue = match[4]
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

func RemoveDisk(params *structs.RmdiskParams) error {
	path := params.Path
	path = strings.Trim(path, "\"'")

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el disco no existe en la ruta especificada: %s", path)
		}
		return fmt.Errorf("error al verificar si el disco existe: %v", err)
	}

	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("error al eliminar el disco: %v", err)
	}

	fmt.Printf("Disco eliminado exitosamente: %s\n", path)
	return nil
}
