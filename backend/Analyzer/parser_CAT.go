package Analyzer

import (
	"errors"
	"regexp"
)

// CatParams representa los parámetros del comando CAT
type CatParams struct {
	Files []string // Lista de archivos a mostrar
}

// ParseCat parsea el comando CAT y extrae los parámetros
func ParseCat(command string) (*CatParams, error) {
	// Expresión regular para capturar parámetros -filen con valores entre comillas o sin comillas
	re := regexp.MustCompile(`-file(\d+)=(?:"([^"]+)"|([^\s]+))`)
	matches := re.FindAllStringSubmatch(command, -1)

	if len(matches) == 0 {
		return nil, errors.New("el comando CAT requiere al menos un parámetro -file1")
	}

	params := &CatParams{
		Files: make([]string, 0),
	}

	// Mapa para mantener orden de archivos por número
	fileMap := make(map[string]string)
	maxFileNum := 0

	for _, match := range matches {
		fileNum := match[1]
		var filePath string
		
		// match[2] es para comillas dobles, match[3] es para sin comillas
		if match[2] != "" {
			filePath = match[2]
		} else {
			filePath = match[3]
		}

		if filePath == "" {
			return nil, errors.New("ruta de archivo vacía en parámetro -file" + fileNum)
		}

		fileMap[fileNum] = filePath
		
		// Convertir fileNum a int para encontrar el máximo
		num := 0
		for i, char := range fileNum {
			if i == 0 && len(fileNum) > 1 {
				// Si hay más de un dígito, multiplicar por 10
				num = int(char-'0') * 10
			} else {
				num += int(char - '0')
			}
		}
		if num > maxFileNum {
			maxFileNum = num
		}
	}

	// Agregar archivos en orden numérico
	for i := 1; i <= maxFileNum; i++ {
		fileNumStr := string(rune(i + '0'))
		if i >= 10 {
			// Para números de dos dígitos
			fileNumStr = string(rune((i/10) + '0')) + string(rune((i%10) + '0'))
		}
		
		if filePath, exists := fileMap[fileNumStr]; exists {
			params.Files = append(params.Files, filePath)
		}
	}

	if len(params.Files) == 0 {
		return nil, errors.New("no se encontraron archivos válidos en los parámetros")
	}

	return params, nil
}
