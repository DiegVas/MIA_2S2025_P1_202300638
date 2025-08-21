package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleMount(w http.ResponseWriter, command string) {
	fmt.Println("handleMount - command:", command)
	params, err := Analyzer.ParseMount(command)
	if err != nil {
		fmt.Printf("handleMount - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	id, err := mounted.MountPartition(params)
	if err != nil {

		response := structs.Response{
			Message: fmt.Sprintf("Error al montar partición: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	confirmationMessage := fmt.Sprintf("Partición activada y montada exitosamente en memoria \n Name: %s, ID: %s, \n Path: %s,\n",
		params.Name, id, params.Path)

	response := structs.Response{
		Message: confirmationMessage,
		Data:    params,
		Type:    "system",
	}

	utils.SendJSONResponse(w, response, http.StatusOK)
}
