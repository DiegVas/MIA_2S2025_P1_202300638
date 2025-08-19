package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

func ParseMkgrp(command string) (*structs.GroupParams, error) {

	params := &structs.GroupParams{}

	session := structs.GetSession()
	if !session.Active {
		return nil, errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}

	if session.User != "root" {
		return nil, errors.New("solo el usuario root puede crear grupos")
	}

	commandLower := strings.ToLower(command)

	if !strings.HasPrefix(commandLower, "mkgrp") {
		return nil, errors.New("el comando debe comenzar con 'mkgrp'")
	}

	paramStr := ""
	if strings.HasPrefix(commandLower, "mkgrp ") {
		paramStr = command[6:]
	} else {
		paramStr = command[5:]
	}
	paramStr = strings.TrimSpace(paramStr)

	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	requiredParams := map[string]bool{
		"name": false,
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
		case "name":
			params.Name = paramValue
			requiredParams["name"] = true
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
func ParseRmgrp(command string) (*structs.GroupParams, error) {

	params := &structs.GroupParams{}

	session := structs.GetSession()
	if !session.Active {
		return nil, errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}

	if session.User != "root" {
		return nil, errors.New("solo el usuario root puede eliminar grupos")
	}

	commandLower := strings.ToLower(command)

	if !strings.HasPrefix(commandLower, "rmgrp") {
		return nil, errors.New("el comando debe comenzar con 'rmgrp'")
	}

	paramStr := ""
	if strings.HasPrefix(commandLower, "rmgrp ") {
		paramStr = command[6:]
	} else {
		paramStr = command[5:]
	}
	paramStr = strings.TrimSpace(paramStr)

	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)

	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}

	requiredParams := map[string]bool{
		"name": false,
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
		case "name":
			params.Name = paramValue
			requiredParams["name"] = true
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
