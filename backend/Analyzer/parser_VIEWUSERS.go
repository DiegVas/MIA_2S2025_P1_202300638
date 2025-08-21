package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

func ParseViewUsers(command string) (*structs.ViewUsersParams, error) {
	params := &structs.ViewUsersParams{}

	// Verificar que hay una sesión activa
	session := structs.GetSession()
	if !session.Active {
		return nil, errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}

	// Solo el usuario root puede ejecutar este comando
	if session.User != "root" {
		return nil, errors.New("solo el usuario root puede ver la información de usuarios y grupos")
	}

	commandLower := strings.ToLower(command)

	if !strings.HasPrefix(commandLower, "viewusers") {
		return nil, errors.New("el comando debe comenzar con 'viewusers'")
	}

	paramStr := ""
	if strings.HasPrefix(commandLower, "viewusers ") {
		paramStr = command[10:]
	} else {
		paramStr = command[9:]
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
