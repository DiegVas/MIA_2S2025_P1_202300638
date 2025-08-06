package routes

import (
	"encoding/json"
	"fmt"
	"main/Analyzer"
	"main/structs"
	"net/http"
)

// handleMkgrp procesa el comando MKGRP para crear un nuevo grupo
func handleMkgrp(w http.ResponseWriter, command string) {
	// Parsear el comando MKGRP
	params, err := Analyzer.ParseMkgrp(command)
	if err != nil {
		fmt.Printf("handleMkgrp - error: %s\n", err.Error())

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

	// Crear el grupo
	err = Analyzer.CreateGroup(params)
	if err != nil {
		fmt.Printf("handleMkgrp - error: %s\n", err.Error())

		// Si hay error al crear el grupo
		response := structs.Response{
			Message: fmt.Sprintf("Error al crear grupo: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y el grupo se creó correctamente
	confirmationMessage := fmt.Sprintf("Grupo '%s' creado exitosamente", params.Name)

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

// handleRmgrp procesa el comando RMGRP para eliminar un grupo
func handleRmgrp(w http.ResponseWriter, command string) {
	// Parsear el comando RMGRP
	params, err := Analyzer.ParseRmgrp(command)
	if err != nil {
		fmt.Printf("handleRmgrp - error: %s\n", err.Error())

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

	// Eliminar el grupo
	err = Analyzer.RemoveGroup(params)
	if err != nil {
		fmt.Printf("handleRmgrp - error: %s\n", err.Error())

		// Si hay error al eliminar el grupo
		response := structs.Response{
			Message: fmt.Sprintf("Error al eliminar grupo: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y el grupo se eliminó correctamente
	confirmationMessage := fmt.Sprintf("Grupo '%s' eliminado exitosamente", params.Name)

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
