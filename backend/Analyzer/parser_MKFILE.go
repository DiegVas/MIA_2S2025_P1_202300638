package Analyzer

import (
	"errors"
	"main/structs"
	"regexp"
	"strconv"
)

func ParseMkfile(command string) (*structs.MkfileParams, error) {
	params := &structs.MkfileParams{
		Size:    0,
		SizeSet: false,
		ContSet: false,
	}

	pathRegex := regexp.MustCompile(`-path=(?:"([^"]+)"|([^\s]+))`)
	sizeRegex := regexp.MustCompile(`-size=([0-9]+)`)
	contRegex := regexp.MustCompile(`-cont=(?:"([^"]+)"|([^\s]+))`)
	rRegex := regexp.MustCompile(`\s-r(?:\s|$)`)

	pathMatch := pathRegex.FindStringSubmatch(command)
	if pathMatch == nil {
		return nil, errors.New("el parámetro -path es obligatorio")
	}

	if pathMatch[1] != "" {
		params.Path = pathMatch[1]
	} else {
		params.Path = pathMatch[2]
	}

	if params.Path == "" {
		return nil, errors.New("el parámetro -path no puede estar vacío")
	}

	if rRegex.MatchString(command) {
		params.R = true

		rWithValueRegex := regexp.MustCompile(`-r=`)
		if rWithValueRegex.MatchString(command) {
			return nil, errors.New("el parámetro -r no debe tener valor")
		}
	}

	sizeMatch := sizeRegex.FindStringSubmatch(command)
	if sizeMatch != nil {
		params.SizeSet = true
		size, err := strconv.Atoi(sizeMatch[1])
		if err != nil {
			return nil, errors.New("el parámetro -size debe ser un número válido")
		}
		if size < 0 {
			return nil, errors.New("el parámetro -size no puede ser negativo")
		}
		params.Size = int32(size)
	}

	contMatch := contRegex.FindStringSubmatch(command)
	if contMatch != nil {
		params.ContSet = true
		if contMatch[1] != "" {
			params.Cont = contMatch[1]
		} else {
			params.Cont = contMatch[2]
		}

		if params.Cont == "" {
			return nil, errors.New("el parámetro -cont no puede estar vacío")
		}
	}

	return params, nil
}
