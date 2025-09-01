package main

import (
	"fmt"
	"main/utils"
	"os"
)

// Función de ejemplo para verificar EBRs en un disco específico
func testEBR() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run test_ebr.go <ruta_del_disco>")
		fmt.Println("Ejemplo: go run test_ebr.go /tmp/disk1.dsk")
		return
	}

	diskPath := os.Args[1]

	// Verificar que el archivo existe
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo %s no existe\n", diskPath)
		return
	}

	// Primero verificar la estructura EBR
	utils.TestEBRStructureSize()

	// Solicitar la posición de inicio de la partición extendida
	// En un caso real, esto se obtendría leyendo el MBR
	var extendedStart int32
	fmt.Print("\nIngrese la posición de inicio de la partición extendida (o 0 para usar valor por defecto): ")
	_, err := fmt.Scanf("%d", &extendedStart)
	if err != nil || extendedStart == 0 {
		extendedStart = 47 + 59*4 // Posición típica después del MBR y particiones
		fmt.Printf("Usando posición por defecto: %d\n", extendedStart)
	}

	// Verificar los EBRs
	fmt.Printf("\nVerificando EBRs en disco: %s\n", diskPath)
	err = utils.VerifyEBRCreation(diskPath, extendedStart)
	if err != nil {
		fmt.Printf("Error durante la verificación: %v\n", err)
	}
}
