package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

func ParseMkusr(command string) (*structs.UserParams, error) {
	params := &structs.UserParams{}
	commandLower := strings.ToLower(command)
	if !strings.HasPrefix(commandLower, "mkusr") {
		return nil, errors.New("el comando debe comenzar con 'mkusr'")
	}
	paramStr := command[5:]
	paramStr = strings.TrimSpace(paramStr)
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	required := map[string]bool{"user": false, "pass": false, "grp": false}
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
		case "grp":
			params.Grp = value
			required["grp"] = true
		default:
			return nil, fmt.Errorf("parámetro desconocido: '%s'", name)
		}
	}
	for k, v := range required {
		if !v {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", k)
		}
	}
	if len(params.User) > 10 || len(params.Pass) > 10 || len(params.Grp) > 10 {
		return nil, errors.New("los parámetros -user, -pass y -grp deben tener máximo 10 caracteres")
	}
	return params, nil
}

func ParseRmusr(command string) (*structs.RmUserParams, error) {
	params := &structs.RmUserParams{}
	commandLower := strings.ToLower(command)
	if !strings.HasPrefix(commandLower, "rmusr") {
		return nil, errors.New("el comando debe comenzar con 'rmusr'")
	}
	paramStr := command[5:]
	paramStr = strings.TrimSpace(paramStr)
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	required := map[string]bool{"user": false}
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

func ParseChgrp(command string) (*structs.ChgrpParams, error) {
	params := &structs.ChgrpParams{}
	commandLower := strings.ToLower(command)
	if !strings.HasPrefix(commandLower, "chgrp") {
		return nil, errors.New("el comando debe comenzar con 'chgrp'")
	}
	paramStr := command[5:]
	paramStr = strings.TrimSpace(paramStr)
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	required := map[string]bool{"user": false, "grp": false}
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
		case "grp":
			params.Grp = value
			required["grp"] = true
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
