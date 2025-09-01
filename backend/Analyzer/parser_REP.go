package Analyzer

import (
	"fmt"
	"main/structs"
	"regexp"
	"strings"
)

// ParseRep parsea el comando REP y extrae sus parámetros
func ParseRep(input string) (*structs.RepParams, error) {

	params := &structs.RepParams{}

	// Validar que el comando sea REP
	if !strings.HasPrefix(strings.ToUpper(input), "REP") {
		fmt.Printf("ERROR: No comienza con REP\n")
		return nil, fmt.Errorf("el comando debe comenzar con REP")
	}

	// Expresión regular para capturar parámetros con o sin comillas
	re := regexp.MustCompile(`-(\w+)=(?:"([^"]*)"|([^\s]+))`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		fmt.Println("ERROR: No se encontraron parámetros válidos")
		return nil, fmt.Errorf("no se encontraron parámetros válidos")
	}

	for _, match := range matches {
		param := strings.ToLower(match[1])
		var value string
		if match[2] != "" {
			value = match[2] // Valor entre comillas
		} else {
			value = match[3] // Valor sin comillas
		}

		switch param {
		case "name":
			// Validar nombres de reporte válidos
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "tree", "sb", "file", "ls"}
			isValid := false
			for _, validName := range validNames {
				if strings.ToLower(value) == validName {
					params.Name = strings.ToLower(value)
					isValid = true
					break
				}
			}
			if !isValid {
				return nil, fmt.Errorf("nombre de reporte inválido: %s. Valores válidos: %v", value, validNames)
			}
		case "path":
			if value == "" {
				return nil, fmt.Errorf("el parámetro path no puede estar vacío")
			}
			params.Path = value
		case "id":
			if value == "" {
				return nil, fmt.Errorf("el parámetro id no puede estar vacío")
			}
			params.ID = value
		case "path_file_ls":
			params.PathFileLS = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", param)
		}
	}

	// Validar parámetros obligatorios
	if params.Name == "" {
		return nil, fmt.Errorf("el parámetro -name es obligatorio")
	}
	if params.Path == "" {
		return nil, fmt.Errorf("el parámetro -path es obligatorio")
	}
	if params.ID == "" {
		return nil, fmt.Errorf("el parámetro -id es obligatorio")
	}

	// Validar que para reportes file y ls se proporcione path_file_ls
	if (params.Name == "file" || params.Name == "ls") && params.PathFileLS == "" {
		return nil, fmt.Errorf("para reportes de tipo '%s' el parámetro -path_file_ls es obligatorio", params.Name)
	}

	return params, nil
}
