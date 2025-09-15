package reports

import (
	"fmt"
	"main/provider/files"
	"main/structs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func generateFileTextReport(params *structs.RepParams) (string, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Usar la función ReadFileContent existente para leer el contenido del archivo
	// Nota: Esta función requiere que haya una sesión activa y usa la partición montada de la sesión
	fileContent, err := files.ReadFileContent(params.PathFileLS)
	if err != nil {
		return "", fmt.Errorf("error leyendo el archivo '%s': %v", params.PathFileLS, err)
	}

	// Generar contenido del reporte
	var content strings.Builder
	content.WriteString("REPORTE DE ARCHIVO\n")
	content.WriteString("==================\n")
	content.WriteString(fmt.Sprintf("Partición: %s\n", params.ID))
	content.WriteString(fmt.Sprintf("Archivo: %s\n", params.PathFileLS))
	content.WriteString(fmt.Sprintf("Tamano: %d bytes\n", len(fileContent)))

	// Agregar fecha actual como fecha del reporte
	currentTime := time.Now()
	content.WriteString(fmt.Sprintf("Fecha del reporte: %s\n", currentTime.Format("2006-01-02 15:04:05")))

	content.WriteString("\nCONTENIDO DEL ARCHIVO:\n")
	content.WriteString("=====================\n")
	content.WriteString(fileContent)

	// Si el archivo no termina en nueva línea, agregarla
	if !strings.HasSuffix(fileContent, "\n") {
		content.WriteString("\n")
	}

	// Escribir archivo
	err = os.WriteFile(params.Path, []byte(content.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("error escribiendo archivo de reporte: %v", err)
	}

	return params.Path, nil
}
