package reports

import (
	"fmt"
	"main/structs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateReport genera el reporte solicitado usando Graphviz
func GenerateReport(params *structs.RepParams) (string, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar contenido según el tipo de reporte
	var dotContent string
	var err error

	switch params.Name {
	case "mbr":
		dotContent, err = generateMBRReport(params.ID)
	case "disk":
		dotContent, err = generateDiskReport(params.ID)
	case "inode":
		dotContent, err = generateInodeReport(params.ID)
	case "block":
		dotContent, err = generateBlockReport(params.ID)
	case "bm_inode":
		return generateBMInodeTextReport(params)
	case "bm_block":
		return generateBMBlockTextReport(params)
	case "tree":
		dotContent, err = generateTreeReport(params.ID)
	case "sb":
		dotContent, err = generateSuperblockReport(params.ID)
	case "file":
		return generateFileTextReport(params)
	case "ls":
		dotContent, err = generateLSReport(params.ID, params.PathFileLS)
	default:
		return "", fmt.Errorf("tipo de reporte no soportado: %s", params.Name)
	}

	if err != nil {
		return "", fmt.Errorf("error generando contenido del reporte: %v", err)
	}

	// Verificar que el contenido no esté vacío
	if len(strings.TrimSpace(dotContent)) == 0 {
		return "", fmt.Errorf("el contenido DOT generado está vacío")
	}

	// Guardar archivo .dot en /tmp/
	dotPath, err := saveDotFile(dotContent, params)
	if err != nil {
		return "", fmt.Errorf("error guardando archivo .dot: %v", err)
	}

	// Renderizar usando comando dot directo
	outputPath, err := renderWithDotCommand(dotPath, params.Path)
	if err != nil {
		return "", fmt.Errorf("error renderizando con comando dot: %v", err)
	}

	return outputPath, nil
}

// renderWithDotCommand renderiza usando el comando dot directamente
func renderWithDotCommand(dotPath, outputPath string) (string, error) {
	// Determinar formato basado en la extensión del archivo de salida
	ext := strings.ToLower(filepath.Ext(outputPath))
	var format string

	switch ext {
	case ".png":
		format = "png"
	case ".jpg", ".jpeg":
		format = "jpeg"
	case ".pdf":
		format = "pdf"
	case ".svg":
		format = "svg"
	default:
		// Por defecto PNG
		format = "png"
		if ext == "" {
			outputPath += ".png"
		}
	}

	// Crear directorio de salida si no existe
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Ejecutar comando dot
	cmd := exec.Command("dot", fmt.Sprintf("-T%s", format), "-o", outputPath, dotPath)

	// Capturar salida de error para debug
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando comando dot: %v\nSalida: %s", err, string(output))
	}

	return outputPath, nil
}

// saveDotFile guarda el contenido DOT en provider/reports/src y devuelve el path
func saveDotFile(dotContent string, params *structs.RepParams) (string, error) {
	// Usar directorio provider/reports/src
	srcDir := "provider/reports/src"

	// Crear directorio si no existe
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio src: %v", err)
	}

	// Generar nombre del archivo .dot basado en el reporte
	dotFileName := fmt.Sprintf("%s_%s.dot", params.Name, params.ID)
	dotPath := filepath.Join(srcDir, dotFileName)

	// Escribir el contenido DOT al archivo
	if err := os.WriteFile(dotPath, []byte(dotContent), 0644); err != nil {
		return "", fmt.Errorf("error escribiendo archivo .dot: %v", err)
	}

	fmt.Printf("Archivo DOT guardado en: %s\n", dotPath)
	return dotPath, nil
}
