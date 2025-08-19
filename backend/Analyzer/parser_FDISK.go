package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type FdiskParams = structs.FdiskParams

func ParseFdisk(command string) (*FdiskParams, error) {

	params := &FdiskParams{
		Size: 0,
		Unit: 'K',
		Path: "",
		Type: 'P',
		Fit:  "WF",
		Name: "",
	}

	command = strings.TrimSpace(command)
	if !strings.HasPrefix(strings.ToLower(command), "fdisk") {
		return nil, errors.New("comando inválido: debe comenzar con 'fdisk'")
	}

	paramStr := strings.TrimSpace(command[5:]) // Remover "fdisk"

	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	sizeFound := false
	pathFound := false
	nameFound := false

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
		case "size":
			size, err := strconv.Atoi(paramValue)
			if err != nil {
				return nil, fmt.Errorf("error al convertir size: %v", err)
			}
			if size <= 0 {
				return nil, errors.New("el parámetro size debe ser positivo y mayor que cero")
			}
			params.Size = size
			sizeFound = true

		case "unit":
			unit := strings.ToUpper(paramValue)
			if unit != "B" && unit != "K" && unit != "M" {
				return nil, errors.New("el parámetro unit debe ser B, K o M")
			}
			params.Unit = unit[0]

		case "path":
			if paramValue == "" {
				return nil, errors.New("el parámetro path no puede estar vacío")
			}
			// Verificar que el disco exista
			path := strings.Trim(paramValue, "\"'")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return nil, fmt.Errorf("el disco en la ruta '%s' no existe", path)
			}
			params.Path = path
			pathFound = true

		case "type":
			typeParam := strings.ToUpper(paramValue)
			if typeParam != "P" && typeParam != "E" && typeParam != "L" {
				return nil, errors.New("el parámetro type debe ser P, E o L")
			}
			params.Type = typeParam[0]

		case "fit":
			fit := strings.ToUpper(paramValue)
			if fit != "BF" && fit != "FF" && fit != "WF" {
				return nil, errors.New("el parámetro fit debe ser BF, FF o WF")
			}
			params.Fit = fit

		case "name":
			if paramValue == "" {
				return nil, errors.New("el parámetro name no puede estar vacío")
			}
			params.Name = paramValue
			nameFound = true

		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", paramName)
		}
	}

	if !pathFound {
		return nil, errors.New("el parámetro -path es obligatorio")
	}
	if !nameFound {
		return nil, errors.New("el parámetro -name es obligatorio")
	}
	if !sizeFound {
		return nil, errors.New("el parámetro -size es obligatorio para crear una partición")
	}

	return params, nil
}
