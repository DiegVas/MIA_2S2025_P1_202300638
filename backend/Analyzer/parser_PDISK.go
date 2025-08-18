package Analyzer

import (
	"errors"
	"regexp"
	"strings"
)

type PdiskParams struct {
	Path string
}

func ParsePdisk(command string) (*PdiskParams, error) {

	re := regexp.MustCompile(`-path=(?:"([^"]+)"|([^\s]+))`)
	matches := re.FindStringSubmatch(command)

	if len(matches) == 0 {
		return nil, errors.New("el comando PDISK requiere el parámetro -path")
	}

	params := &PdiskParams{}

	if matches[1] != "" {
		params.Path = matches[1]
	} else {
		params.Path = matches[2]
	}

	if params.Path == "" {
		return nil, errors.New("la ruta del disco no puede estar vacía")
	}

	if strings.Contains(params.Path, "..") {
		return nil, errors.New("la ruta no puede contener '..'")
	}

	return params, nil
}
