package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strconv"
	"strings"
)

type MkdiskParams = structs.MkdiskParams

func ParseMkdisk(command string) (*MkdiskParams, error) {

	// ! Parametros iniciales

	params := &MkdiskParams{
		Size: 0,
		Fit:  "FF",
		Unit: 'M',
		Path: "",
	}

	command = strings.TrimSpace(command)
	if !strings.HasPrefix(strings.ToLower(command), "mkdisk") {
		return nil, errors.New("comando inválido: debe comenzar con 'mkdisk'")
	}

	// ? Obtener solo los parámetros
	paramStr := strings.TrimSpace(command[6:]) // Remover "mkdisk"

	// TODO: Captura: -parámetro="valor con espacios" o -parámetro='valor con espacios' o -parámetro=valor_sin_espacios
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	sizeFound := false
	pathFound := false

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		paramName := strings.ToLower(match[1])

		var paramValue string
		if match[2] != "" {
			paramValue = match[2] // ? Valor con comillas dobles
		} else if match[3] != "" {
			paramValue = match[3] // ? Valor con comillas simples
		} else {
			paramValue = match[4] // ? Valor sin comillas
		}

		// * Parametros

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

		case "fit":
			fit := strings.ToUpper(paramValue)
			if fit != "BF" && fit != "FF" && fit != "WF" {
				return nil, errors.New("el parámetro fit debe ser BF, FF o WF")
			}
			params.Fit = fit

		case "unit":
			unit := strings.ToUpper(paramValue)
			if unit != "K" && unit != "M" {
				return nil, errors.New("el parámetro unit debe ser K o M")
			}
			params.Unit = unit[0]

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

	if !sizeFound {
		return nil, errors.New("el parámetro -size es obligatorio")
	}
	if !pathFound {
		return nil, errors.New("el parámetro -path es obligatorio")
	}

	return params, nil
}
