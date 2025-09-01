package utils

import (
	"fmt"
	"main/structs"
	"os"
)

// VerifyEBRCreation verifica que los EBRs se estén creando correctamente
func VerifyEBRCreation(diskPath string, extendedPartitionStart int32) error {
	fmt.Println("=== VERIFICACIÓN DE EBRs ===")

	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	currentPos := extendedPartitionStart
	ebrCount := 0
	ebrSize := int32(30)

	fmt.Printf("Inicio de verificación en posición: %d\n", extendedPartitionStart)

	for {
		// Leer el EBR
		ebrData := make([]byte, ebrSize)
		_, err := file.ReadAt(ebrData, int64(currentPos))
		if err != nil {
			fmt.Printf("Error leyendo EBR en posición %d: %v\n", currentPos, err)
			break
		}

		// Extraer campos del EBR
		partMount := ebrData[0]
		partFit := ebrData[1]
		partStart := int32(ebrData[2]) | int32(ebrData[3])<<8 | int32(ebrData[4])<<16 | int32(ebrData[5])<<24
		partSize := int32(ebrData[6]) | int32(ebrData[7])<<8 | int32(ebrData[8])<<16 | int32(ebrData[9])<<24
		partNext := int32(ebrData[10]) | int32(ebrData[11])<<8 | int32(ebrData[12])<<16 | int32(ebrData[13])<<24
		partName := string(ebrData[14:30])

		// Limpiar el nombre
		var cleanName string
		for i := 0; i < len(partName); i++ {
			if partName[i] == 0 {
				break
			}
			cleanName += string(partName[i])
		}

		// Si no hay partición, terminar
		if partSize == 0 {
			fmt.Printf("EBR vacío encontrado en posición %d\n", currentPos)
			break
		}

		ebrCount++
		fmt.Printf("\n--- EBR #%d ---\n", ebrCount)
		fmt.Printf("Posición EBR: %d\n", currentPos)
		fmt.Printf("Part_mount: %d\n", partMount)
		fmt.Printf("Part_fit: %c\n", partFit)
		fmt.Printf("Part_start: %d (posición absoluta: %d)\n", partStart, currentPos+partStart)
		fmt.Printf("Part_size: %d bytes\n", partSize)
		fmt.Printf("Part_next: %d", partNext)
		if partNext != 0 {
			fmt.Printf(" (próximo EBR en: %d)\n", extendedPartitionStart+partNext)
		} else {
			fmt.Printf(" (último EBR)\n")
		}
		fmt.Printf("Part_name: '%s'\n", cleanName)

		// Verificaciones de integridad
		fmt.Printf("\n--- Verificaciones ---\n")

		// 1. Verificar que Part_start sea >= tamaño del EBR
		if partStart < ebrSize {
			fmt.Printf("⚠️  ADVERTENCIA: Part_start (%d) es menor que el tamaño del EBR (%d)\n", partStart, ebrSize)
		} else {
			fmt.Printf("✅ Part_start es válido\n")
		}

		// 2. Verificar que el nombre no esté vacío
		if cleanName == "" {
			fmt.Printf("⚠️  ADVERTENCIA: Part_name está vacío\n")
		} else {
			fmt.Printf("✅ Part_name es válido: '%s'\n", cleanName)
		}

		// 3. Verificar fit válido
		if partFit == 'F' || partFit == 'B' || partFit == 'W' {
			fmt.Printf("✅ Part_fit es válido: %c\n", partFit)
		} else {
			fmt.Printf("⚠️  ADVERTENCIA: Part_fit inválido: %c\n", partFit)
		}

		// Si no hay siguiente EBR, terminar
		if partNext == 0 {
			fmt.Printf("\n✅ Última partición lógica alcanzada\n")
			break
		}

		// Ir al siguiente EBR
		nextPos := extendedPartitionStart + partNext

		// Verificar que la posición sea válida
		if nextPos <= currentPos {
			fmt.Printf("❌ ERROR: Posición del siguiente EBR (%d) es inválida (actual: %d)\n", nextPos, currentPos)
			break
		}

		currentPos = nextPos
	}

	fmt.Printf("\n=== RESUMEN ===\n")
	fmt.Printf("Total de particiones lógicas encontradas: %d\n", ebrCount)

	if ebrCount == 0 {
		fmt.Printf("❌ No se encontraron particiones lógicas\n")
	} else {
		fmt.Printf("✅ EBRs encontrados y verificados\n")
	}

	return nil
}

// TestEBRStructureSize verifica que el tamaño de la estructura EBR sea correcto
func TestEBRStructureSize() {
	fmt.Println("\n=== VERIFICACIÓN DEL TAMAÑO DE ESTRUCTURA EBR ===")

	var ebr structs.EBR

	// Calcular tamaño esperado
	expectedSize := 1 + 1 + 4 + 4 + 4 + 16 // 30 bytes
	fmt.Printf("Tamaño esperado de EBR: %d bytes\n", expectedSize)

	// Verificar campos individuales
	fmt.Printf("Part_mount: %d bytes\n", len(ebr.Part_mount))
	fmt.Printf("Part_fit: %d bytes\n", len(ebr.Part_fit))
	fmt.Printf("Part_start: 4 bytes (int32)\n")
	fmt.Printf("Part_size: 4 bytes (int32)\n")
	fmt.Printf("Part_next: 4 bytes (int32)\n")
	fmt.Printf("Part_name: %d bytes\n", len(ebr.Part_name))

	total := len(ebr.Part_mount) + len(ebr.Part_fit) + 4 + 4 + 4 + len(ebr.Part_name)
	fmt.Printf("Total calculado: %d bytes\n", total)

	if total == expectedSize {
		fmt.Printf("✅ Estructura EBR tiene el tamaño correcto\n")
	} else {
		fmt.Printf("❌ ERROR: Estructura EBR tiene tamaño incorrecto\n")
	}
}
