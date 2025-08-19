package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleMkfs(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkfs(command)

	if err != nil {
		fmt.Printf("handleMkfs - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = mounted.FormatPartition(params)
	if err != nil {
		response := structs.Response{
			Message: fmt.Sprintf("Error al formatear la partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := structs.Response{
		Message: fmt.Sprintf("Partición formateada exitosamente en sistema de archivos ext2 - ID: %s, Type: %s",
			params.ID, params.Type),
		Data: params,
		Type: "system",
	}

	utils.SendJSONResponse(w, response, http.StatusOK)
}
