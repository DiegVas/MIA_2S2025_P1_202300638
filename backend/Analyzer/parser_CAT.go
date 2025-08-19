package Analyzer

import (
	"errors"
	"regexp"
)

type CatParams struct {
	Files []string
}

func ParseCat(command string) (*CatParams, error) {

	re := regexp.MustCompile(`-file(\d+)=(?:"([^"]+)"|([^\s]+))`)
	matches := re.FindAllStringSubmatch(command, -1)

	if len(matches) == 0 {
		return nil, errors.New("el comando CAT requiere al menos un parámetro -file1")
	}

	params := &CatParams{
		Files: make([]string, 0),
	}

	fileMap := make(map[string]string)
	maxFileNum := 0

	for _, match := range matches {
		fileNum := match[1]
		var filePath string

		if match[2] != "" {
			filePath = match[2]
		} else {
			filePath = match[3]
		}

		if filePath == "" {
			return nil, errors.New("ruta de archivo vacía en parámetro -file" + fileNum)
		}

		fileMap[fileNum] = filePath

		num := 0
		for i, char := range fileNum {
			if i == 0 && len(fileNum) > 1 {

				num = int(char-'0') * 10
			} else {
				num += int(char - '0')
			}
		}
		if num > maxFileNum {
			maxFileNum = num
		}
	}

	for i := 1; i <= maxFileNum; i++ {
		fileNumStr := string(rune(i + '0'))
		if i >= 10 {

			fileNumStr = string(rune((i/10)+'0')) + string(rune((i%10)+'0'))
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
