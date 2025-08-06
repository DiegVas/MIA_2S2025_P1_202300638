package Analyzer

import (
	"errors"
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

// GroupParams estructura para los parámetros de MKGRP y RMGRP
type GroupParams struct {
	Name string
}

// ParseMkgrp analiza el comando MKGRP y extrae los parámetros
func ParseMkgrp(command string) (*GroupParams, error) {
	// Crear la estructura para almacenar los parámetros
	params := &GroupParams{}
	
	// Verificar la sesión activa (solo root puede ejecutar este comando)
	session := structs.GetSession()
	if !session.Active {
		return nil, errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	
	if session.User != "root" {
		return nil, errors.New("solo el usuario root puede crear grupos")
	}
	
	// Convertir a minúsculas solo para verificar el prefijo del comando
	commandLower := strings.ToLower(command)
	
	// Verificar que el comando comience con "mkgrp"
	if !strings.HasPrefix(commandLower, "mkgrp") {
		return nil, errors.New("el comando debe comenzar con 'mkgrp'")
	}

	// Obtener solo los parámetros - usamos el comando original para mantener case-sensitivity
	paramStr := ""
	if strings.HasPrefix(commandLower, "mkgrp ") {
		paramStr = command[6:] // Remover "mkgrp "
	} else {
		paramStr = command[5:] // Remover "mkgrp" sin espacio
	}
	paramStr = strings.TrimSpace(paramStr)

	// Expresión regular para capturar parámetros con valores que pueden contener espacios
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	
	// Mapa para verificar parámetros obligatorios
	requiredParams := map[string]bool{
		"name": false,
	}
	
	// Procesar los parámetros encontrados
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		// El nombre del parámetro se convierte a minúsculas para ser insensible a mayúsculas/minúsculas
		paramName := strings.ToLower(match[1])
		
		// Determinar el valor del parámetro dependiendo de qué grupo de captura contiene el valor
		// NO convertimos el valor a minúsculas para mantener sensibilidad a mayúsculas/minúsculas
		var paramValue string
		if match[2] != "" {
			paramValue = match[2] // Valor con comillas dobles
		} else if match[3] != "" {
			paramValue = match[3] // Valor con comillas simples
		} else if match[4] != "" {
			paramValue = match[4] // Valor sin comillas
		}
		
		switch paramName {
		case "name":
			params.Name = paramValue // Mantener mayúsculas/minúsculas originales
			requiredParams["name"] = true
		default:
			return nil, fmt.Errorf("parámetro desconocido: '%s'", paramName)
		}
	}
	
	// Verificar que todos los parámetros obligatorios estén presentes
	for param, found := range requiredParams {
		if !found {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", param)
		}
	}
	
	return params, nil
}

// ParseRmgrp analiza el comando RMGRP y extrae los parámetros
func ParseRmgrp(command string) (*GroupParams, error) {
	// Crear la estructura para almacenar los parámetros
	params := &GroupParams{}
	
	// Verificar la sesión activa (solo root puede ejecutar este comando)
	session := structs.GetSession()
	if !session.Active {
		return nil, errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	
	if session.User != "root" {
		return nil, errors.New("solo el usuario root puede eliminar grupos")
	}
	
	// Convertir a minúsculas solo para verificar el prefijo del comando
	commandLower := strings.ToLower(command)
	
	// Verificar que el comando comience con "rmgrp"
	if !strings.HasPrefix(commandLower, "rmgrp") {
		return nil, errors.New("el comando debe comenzar con 'rmgrp'")
	}

	// Obtener solo los parámetros - usamos el comando original para mantener case-sensitivity
	paramStr := ""
	if strings.HasPrefix(commandLower, "rmgrp ") {
		paramStr = command[6:] // Remover "rmgrp "
	} else {
		paramStr = command[5:] // Remover "rmgrp" sin espacio
	}
	paramStr = strings.TrimSpace(paramStr)

	// Expresión regular para capturar parámetros con valores que pueden contener espacios
	paramRegex := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)
	matches := paramRegex.FindAllStringSubmatch(paramStr, -1)
	
	if len(matches) == 0 {
		return nil, errors.New("no se encontraron parámetros válidos")
	}
	
	// Mapa para verificar parámetros obligatorios
	requiredParams := map[string]bool{
		"name": false,
	}
	
	// Procesar los parámetros encontrados
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		// El nombre del parámetro se convierte a minúsculas para ser insensible a mayúsculas/minúsculas
		paramName := strings.ToLower(match[1])
		
		// Determinar el valor del parámetro dependiendo de qué grupo de captura contiene el valor
		// NO convertimos el valor a minúsculas para mantener sensibilidad a mayúsculas/minúsculas
		var paramValue string
		if match[2] != "" {
			paramValue = match[2] // Valor con comillas dobles
		} else if match[3] != "" {
			paramValue = match[3] // Valor con comillas simples
		} else if match[4] != "" {
			paramValue = match[4] // Valor sin comillas
		}
		
		switch paramName {
		case "name":
			params.Name = paramValue // Mantener mayúsculas/minúsculas originales
			requiredParams["name"] = true
		default:
			return nil, fmt.Errorf("parámetro desconocido: '%s'", paramName)
		}
	}
	
	// Verificar que todos los parámetros obligatorios estén presentes
	for param, found := range requiredParams {
		if !found {
			return nil, fmt.Errorf("falta el parámetro obligatorio: '%s'", param)
		}
	}
	
	return params, nil
}
