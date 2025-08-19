package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

func ParseLogin(command string) (*structs.LoginParams, error) {
	params := &structs.LoginParams{}
	commandLower := strings.ToLower(command)
	if !strings.HasPrefix(commandLower, "login") {
		return nil, errors.New("el comando debe comenzar con 'login'")
	}
	paramStr := command[5:]
	paramStr = strings.TrimSpace(paramStr)
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	required := map[string]bool{"user": false, "pass": false, "id": false}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := strings.ToLower(match[1])
		var value string
		if match[2] != "" {
			value = match[2]
		} else if match[3] != "" {
			value = match[3]
		} else if match[4] != "" {
			value = match[4]
		}
		switch name {
		case "user":
			params.User = value
			required["user"] = true
		case "pass":
			params.Pass = value
			required["pass"] = true
		case "id":
			params.ID = value
			required["id"] = true
		default:
			return nil, fmt.Errorf("parámetro desconocido: '%s'", name)
		}
	}
	for k, v := range required {
		if !v {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", k)
		}
	}
	return params, nil
}

// ParseLogout analiza el comando LOGOUT
func ParseLogout(command string) error {
	if strings.TrimSpace(strings.ToLower(command)) != "logout" {
		return errors.New("comando inválido: debe ser exactamente 'logout'")
	}
	return nil
}
