package routes

import (
	"encoding/json"
	"fmt"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleRmdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseRmdisk(command)
	if err != nil {
		fmt.Printf("handleRmdisk - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = Analyzer.RemoveDisk(params)
	if err != nil {

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

	confirmationMessage := fmt.Sprintf("Disco eliminado exitosamente - Path: %s", params.Path)

	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
