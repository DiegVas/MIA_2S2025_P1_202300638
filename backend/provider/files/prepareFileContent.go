package files

import (
	"fmt"
	"io/ioutil"
	"main/structs"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func prepareFileContent(params *structs.MkfileParams) (string, int32, error) {
	var content string
	var size int32

	// El parámetro -cont tiene prioridad sobre -size
	if params.ContSet {
		// Leer contenido desde archivo local
		localContent, err := ioutil.ReadFile(params.Cont)
		if err != nil {
			return "", 0, fmt.Errorf("error al leer el archivo local '%s': %v", params.Cont, err)
		}
		content = string(localContent)
		size = int32(len(content))
		fmt.Printf("🎯 DEBUG prepareFileContent: Contenido desde archivo, tamaño: %d\n", size)
	} else if params.SizeSet {
		// Generar contenido con números del 0 al 9
		content = generateNumberContent(params.Size)
		size = params.Size
		fmt.Printf("🎯 DEBUG prepareFileContent: Contenido generado, tamaño solicitado: %d, tamaño real: %d\n", params.Size, len(content))
		fmt.Printf("🎯 DEBUG prepareFileContent: Primeros 100 chars: %q\n", content[:min(100, len(content))])
	} else {
		// Archivo vacío
		content = ""
		size = 0
		fmt.Printf("🎯 DEBUG prepareFileContent: Archivo vacío\n")
	}

	return content, size, nil
}
