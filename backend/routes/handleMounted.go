package routes

import (
	"encoding/json"
	"fmt"
	"main/Analyzer"
	"main/structs"
	"net/http"
)

// handleMounted procesa el comando MOUNTED y devuelve todas las particiones montadas
func handleMounted(w http.ResponseWriter, command string) {
	valid, err := Analyzer.ParseMounted(command)
	if err != nil {
		fmt.Printf("handleMounted - error: %s\n", err.Error())

		// Si hay error en el parsing, enviar mensaje de error
		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !valid {
		// Si el comando no es válido
		response := structs.Response{
			Message: "Comando inválido",
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Obtener las particiones montadas
	partitions := Analyzer.GetMountedPartitions()
	
	// Formatear la respuesta
	formattedOutput := Analyzer.FormatMountedPartitions()

	// Responder con JSON de éxito
	response := structs.Response{
		Message: formattedOutput,
		Data:    partitions,
		Type:    "mounted-partitions", // Tipo específico para particiones montadas
	}

	// Configurar headers de respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
