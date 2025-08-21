package reports

import (
	"bytes"
	"context"
	"fmt"
	"main/structs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-graphviz"
)

// GenerateReport genera el reporte solicitado usando Graphviz
func GenerateReport(params *structs.RepParams) (string, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar contenido DOT según el tipo de reporte
	var dotContent string
	var err error
	switch params.Name {
	case "mbr":
		dotContent, err = generateMBRReport(params.ID)
	case "test":
		// Reporte de prueba simple
		dotContent = `digraph Test_Report {
    rankdir=TB;
    node [shape=box, style=filled];
    
    start [label="Test Report", fillcolor=lightblue];
    middle [label="This is a test\\nto verify Graphviz\\nworking correctly", fillcolor=lightgreen];
    end [label="Success!", fillcolor=lightyellow];
    
    start -> middle -> end;
}`
		err = nil
	case "disk":
		dotContent, err = generateDiskReport(params.ID)
	case "inode":
		dotContent, err = generateInodeReport(params.ID)
	case "block":
		dotContent, err = generateBlockReport(params.ID)
	case "bm_inode":
		dotContent, err = generateBMInodeReport(params.ID)
	case "bm_block":
		dotContent, err = generateBMBlockReport(params.ID)
	case "tree":
		dotContent, err = generateTreeReport(params.ID)
	case "sb":
		dotContent, err = generateSuperblockReport(params.ID)
	case "file":
		dotContent, err = generateFileReport(params.ID, params.PathFileLS)
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

	// Generar imagen con Graphviz
	outputPath, err := renderGraphviz(dotContent, params.Path)
	if err != nil {
		return "", fmt.Errorf("error renderizando con Graphviz: %v", err)
	}

	return outputPath, nil
}

// renderGraphviz renderiza el contenido DOT como imagen
func renderGraphviz(dotContent, outputPath string) (string, error) {
	// Verificar que el contenido DOT no esté vacío
	if len(strings.TrimSpace(dotContent)) == 0 {
		return "", fmt.Errorf("contenido DOT está vacío")
	}

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		return "", fmt.Errorf("error creando instancia de graphviz: %v", err)
	}
	defer g.Close()

	// Parsear el contenido DOT
	graph, err := graphviz.ParseBytes([]byte(dotContent))
	if err != nil {
		return "", fmt.Errorf("error parseando DOT: %v", err)
	}
	defer graph.Close()

	// Determinar formato basado en la extensión
	ext := strings.ToLower(filepath.Ext(outputPath))
	var format graphviz.Format

	switch ext {
	case ".png":
		format = graphviz.PNG
	case ".jpg", ".jpeg":
		format = graphviz.JPG
	case ".svg":
		format = graphviz.SVG
	default:
		// Por defecto PNG
		format = graphviz.PNG
		if ext == "" {
			outputPath += ".png"
		}
	}

	// Renderizar a buffer y escribir al archivo
	var buf bytes.Buffer
	if err := g.Render(ctx, graph, format, &buf); err != nil {
		return "", fmt.Errorf("error renderizando en buffer: %v", err)
	}

	// Escribir el buffer al archivo
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return outputPath, nil
}
