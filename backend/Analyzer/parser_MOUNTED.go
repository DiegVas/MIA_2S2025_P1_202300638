package Analyzer

import (
	"fmt"
	"main/structs"
	"strings"
)

// ParseMounted analiza el comando MOUNTED
// Este comando no tiene parámetros, solo verifica que sea el comando correcto
func ParseMounted(command string) (bool, error) {
	// Convertir a minúsculas y eliminar espacios adicionales
	command = strings.TrimSpace(strings.ToLower(command))
	
	// Verificar que sea exactamente "mounted"
	if command != "mounted" {
		return false, fmt.Errorf("comando inválido: debe ser exactamente 'mounted'")
	}
	
	return true, nil
}

// GetMountedPartitions obtiene todas las particiones montadas en el sistema
func GetMountedPartitions() []structs.MountedPartition {
	// Obtener el gestor de montajes
	mountManager := structs.GetMountManager()
	
	// Obtener todas las particiones montadas
	return mountManager.GetMountedPartitions()
}

// FormatMountedPartitions devuelve un string formateado con las particiones montadas
func FormatMountedPartitions() string {
	partitions := GetMountedPartitions()
	
	if len(partitions) == 0 {
		return "No hay particiones montadas actualmente en memoria"
	}
	
	// Crear un string con información detallada de las particiones montadas
	var result strings.Builder
	result.WriteString("Particiones montadas en memoria:\n\n ")
	
	for i, partition := range partitions {
		result.WriteString(fmt.Sprintf("%d. ID: %s | Nombre: %s | Ruta: %s\n ", 
			i+1, partition.ID, partition.Name, partition.Path))
	}
	
	
	return result.String()
}
