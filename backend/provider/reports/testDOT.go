package reports

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-graphviz"
)

// TestSimpleDOT prueba con un DOT muy básico
func TestSimpleDOT(outputPath string) error {
	simpleDOT := `digraph Test {
    rankdir=TB;
    node [shape=box];
    
    A [label="Test Node"];
    B [label="Another Node"];
    
    A -> B;
}`

	fmt.Printf("=== TEST SIMPLE DOT ===\n")
	fmt.Printf("DOT Content:\n%s\n", simpleDOT)

	// Crear directorio si no existe
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %v", err)
	}

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		return fmt.Errorf("error creando graphviz: %v", err)
	}
	defer g.Close()

	graph, err := graphviz.ParseBytes([]byte(simpleDOT))
	if err != nil {
		return fmt.Errorf("error parseando DOT: %v", err)
	}
	defer graph.Close()

	// Determinar formato
	ext := strings.ToLower(filepath.Ext(outputPath))
	var format graphviz.Format = graphviz.PNG
	if ext == "" {
		outputPath += ".png"
	}

	fmt.Printf("Renderizando archivo de prueba en: %s\n", outputPath)
	if err := g.RenderFilename(ctx, graph, format, outputPath); err != nil {
		return fmt.Errorf("error renderizando: %v", err)
	}

	// Verificar resultado
	if info, err := os.Stat(outputPath); err == nil {
		fmt.Printf("Archivo de prueba creado: tamaño %d bytes\n", info.Size())
		if info.Size() > 0 {
			fmt.Printf("✅ Graphviz funciona correctamente\n")
		} else {
			fmt.Printf("❌ Archivo creado pero vacío\n")
		}
	} else {
		fmt.Printf("❌ Archivo no encontrado: %v\n", err)
	}

	return nil
}
