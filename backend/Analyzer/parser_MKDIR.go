package Analyzer

import (
	"errors"
	"main/structs"
	"regexp"
	"strings"
)

func ParseMkdir(command string) (*structs.MkdirParams, error) {
	var params structs.MkdirParams

	// Verificar que el comando comience con mkdir
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(command)), "mkdir") {
		return nil, errors.New("comando inválido, debe comenzar con 'mkdir'")
	}

	// Regex para path (con o sin comillas)
	pathRegex := regexp.MustCompile(`-path=(?:"([^"]+)"|([^\s]+))`)
	// Regex para el parámetro -p
	pRegex := regexp.MustCompile(`\s-p(?:\s|$)`)

	// Buscar path
	pathMatch := pathRegex.FindStringSubmatch(command)
	if pathMatch == nil {
		return nil, errors.New("el parámetro -path es obligatorio")
	}

	if pathMatch[1] != "" {
		params.Path = pathMatch[1] // Path con comillas
	} else {
		params.Path = pathMatch[2] // Path sin comillas
	}

	// Buscar parámetro -p
	if pRegex.MatchString(command) {
		params.P = true
	}

	// Validaciones
	if params.Path == "" {
		return nil, errors.New("el parámetro -path no puede estar vacío")
	}

	// Validar que la ruta comience con /
	if !strings.HasPrefix(params.Path, "/") {
		return nil, errors.New("la ruta debe ser absoluta (comenzar con /)")
	}

	return &params, nil
}
