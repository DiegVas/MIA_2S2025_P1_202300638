package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/users"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleMkusr(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkusr(command)
	if err != nil {
		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = users.CreateUser(params)
	if err != nil {
		response := structs.Response{
			Message: err.Error(),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := utils.CreateResponse(
		fmt.Sprintf("Usuario '%s' creado exitosamente", params.User),
		params,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)

}

func handleRmusr(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseRmusr(command)
	if err != nil {

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}
	err = users.RemoveUser(params)
	if err != nil {
		response := structs.Response{
			Message: err.Error(),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}
	response := utils.CreateResponse(
		fmt.Sprintf("Usuario '%s' eliminado exitosamente", params.User),
		params,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}

func handleChgrp(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseChgrp(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = users.ChangeUserGroup(params)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			params,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := utils.CreateResponse(
		fmt.Sprintf("Grupo del usuario '%s' cambiado a '%s' exitosamente", params.User, params.Grp),
		params,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}

func handleViewUsers(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseViewUsers(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	viewUsersResponse, err := users.ViewAllUsers(params)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			params,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Generar mensaje detallado con toda la información
	message := users.GenerateDetailedMessage(viewUsersResponse)

	response := utils.CreateResponse(
		message,
		viewUsersResponse,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}
