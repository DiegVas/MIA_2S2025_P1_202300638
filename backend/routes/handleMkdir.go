package routes

import (
	"main/Analyzer"
	"main/provider/files"
	"main/utils"
	"net/http"
)

func handleMkdir(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkdir(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = files.CreateDirectory(params)
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
		"Directorio creado exitosamente",
		params,
		"success",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}
