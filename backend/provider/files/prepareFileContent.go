package files

import (
	"fmt"
	"io/ioutil"
	"main/structs"
)

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

	} else if params.SizeSet {
		// Generar contenido con números del 0 al 9
		content = generateNumberContent(params.Size)
		size = params.Size

	} else {
		// Archivo vacío
		content = ""
		size = 0

	}

	return content, size, nil
}
