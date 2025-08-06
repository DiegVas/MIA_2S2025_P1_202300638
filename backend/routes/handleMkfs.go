package routes

import (
	"encoding/json"
	"fmt"
	"main/Analyzer"
	"main/structs"
	"net/http"
)

// handleMkfs procesa el comando MKFS para formatear una partición
func handleMkfs(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkfs(command)
	if err != nil {
		fmt.Printf("handleMkfs - error: %s\n", err.Error())

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

	err = Analyzer.FormatPartition(params)
	if err != nil {
		// Si hay error al formatear la partición
		response := structs.Response{
			Message: fmt.Sprintf("Error al formatear la partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y la partición se formateó correctamente
	confirmationMessage := fmt.Sprintf("Partición formateada exitosamente en sistema de archivos ext2 - ID: %s, Type: %s",
		params.ID, params.Type)

	// Responder con JSON de éxito
	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	// Configurar headers de respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
