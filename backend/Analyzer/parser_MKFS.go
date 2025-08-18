package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

func ParseMkfs(command string) (*structs.MkfsParams, error) {
	params := &structs.MkfsParams{
		Type: "full",
	}

	originalCommand := command

	commandLower := strings.ToLower(command)

	if !strings.HasPrefix(commandLower, "mkfs") {
		return nil, errors.New("el comando debe comenzar con 'mkfs'")
	}

	paramStr := ""
	if strings.HasPrefix(commandLower, "mkfs ") {
		paramStr = originalCommand[5:]
	} else {
		paramStr = originalCommand[4:]
	}
	paramStr = strings.TrimSpace(paramStr)

	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	requiredParams := map[string]bool{
		"id": false,
	}

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
		} else if match[4] != "" {
			paramValue = match[4]
		}

		switch paramName {
		case "id":
			params.ID = paramValue
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

	for param, found := range requiredParams {
		if !found {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", param)
		}
	}

	return params, nil
}
