package routes

import (
	"fmt"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleFdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseFdisk(command)
	if err != nil {
		fmt.Printf("handleFdisk - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)

		return
	}

	err = utils.CreatePartition(params)
	if err != nil {

		response := structs.Response{
			Message: fmt.Sprintf("Error al crear partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	confirmationMessage := fmt.Sprintf("Partición creada exitosamente - Name: %s, Size: %d, Type: %c, Path: %s",
		params.Name, params.Size, params.Type, params.Path)

	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	utils.SendJSONResponse(w, response, http.StatusCreated)
}
