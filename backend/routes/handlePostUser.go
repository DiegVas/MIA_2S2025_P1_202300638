package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"net/http"
	"strings"
)

func handleMkdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkdisk(command)
	if err != nil {
		fmt.Printf("handleMkdisk - error: %s\n", err.Error())

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

	err = utils.CreateDisk(params)
	if err != nil {
		// Si hay error al crear el disco
		response := structs.Response{
			Message: fmt.Sprintf("Error al crear disco: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y el disco se creó correctamente
	confirmationMessage := fmt.Sprintf("Disco creado exitosamente - Size: %d, Unit: %c, Fit: %s, Path: %s",
		params.Size, params.Unit, params.Fit, params.Path)

	// Responder con JSON de éxito
	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	// Configurar headers de respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleRmdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseRmdisk(command)
	if err != nil {
		fmt.Printf("handleRmdisk - error: %s\n", err.Error())

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

	err = Analyzer.RemoveDisk(params)
	if err != nil {
		// Si hay error al eliminar el disco
		response := structs.Response{
			Message: fmt.Sprintf("Error al eliminar disco: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y el disco se eliminó correctamente
	confirmationMessage := fmt.Sprintf("Disco eliminado exitosamente - Path: %s", params.Path)

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

func handleFdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseFdisk(command)
	if err != nil {
		fmt.Printf("handleFdisk - error: %s\n", err.Error())

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

	err = utils.CreatePartition(params)
	if err != nil {
		// Si hay error al crear la partición
		response := structs.Response{
			Message: fmt.Sprintf("Error al crear partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y la partición se creó correctamente
	confirmationMessage := fmt.Sprintf("Partición creada exitosamente - Name: %s, Size: %d, Type: %c, Path: %s",
		params.Name, params.Size, params.Type, params.Path)

	// Responder con JSON de éxito
	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	// Configurar headers de respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleMount(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMount(command)
	if err != nil {
		fmt.Printf("handleMount - error: %s\n", err.Error())

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

	id, err := Analyzer.MountPartition(params)
	if err != nil {
		// Si hay error al montar la partición
		response := structs.Response{
			Message: fmt.Sprintf("Error al montar partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Si el comando es válido y la partición se montó correctamente
	confirmationMessage := fmt.Sprintf("Partición activada y montada exitosamente - Name: %s, Path: %s, ID: %s (Carnet: 202300638)",
		params.Name, params.Path, id)

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

func HandlePostUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Leer el body de la petición
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el body", http.StatusBadRequest)
		return
	}

	// Decodificar JSON para comando
	var command structs.Command
	err = json.Unmarshal(body, &command)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Determinar qué tipo de comando es basado en el prefijo
	commandLower := strings.ToLower(command.Command)

	if strings.HasPrefix(commandLower, "mkdisk") {
		// Procesar comando MKDISK
		handleMkdisk(w, command.Command)
	} else if strings.HasPrefix(commandLower, "rmdisk") {
		// Procesar comando RMDISK
		handleRmdisk(w, command.Command)
	} else if strings.HasPrefix(commandLower, "fdisk") {
		// Procesar comando FDISK
		handleFdisk(w, command.Command)
	} else if strings.HasPrefix(commandLower, "mount") {
		// Procesar comando MOUNT
		handleMount(w, command.Command)
	} else {
		// Comando no reconocido
		response := structs.Response{
			Message: "Comando no reconocido. Comandos soportados: MKDISK, RMDISK, FDISK, MOUNT",
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}
}
