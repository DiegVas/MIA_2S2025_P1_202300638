package routes

import (
	"fmt"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleMkdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkdisk(command)

	if err != nil {
		fmt.Printf("handleMkdisk - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	// ? Crear disco
	err = utils.CreateDisk(params)

	if err != nil {
		response := structs.Response{
			Message: fmt.Sprintf("Error al crear disco: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	confirmationMessage := fmt.Sprintf("Disco creado exitosamente - Size: %d, Unit: %c, Fit: %s, Path: %s",
		params.Size, params.Unit, params.Fit, params.Path)

	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	utils.SendJSONResponse(w, response, http.StatusCreated)
}
