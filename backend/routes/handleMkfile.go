package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/files"
	"main/utils"
	"net/http"
)

func handleMkfile(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseMkfile(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	err = files.CreateFile(params)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			params,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Construir mensaje de confirmación
	var message string
	if params.ContSet {
		message = fmt.Sprintf("Archivo '%s' creado exitosamente con contenido desde '%s'", params.Path, params.Cont)
	} else if params.SizeSet {
		message = fmt.Sprintf("Archivo '%s' creado exitosamente con tamaño de %d bytes", params.Path, params.Size)
	} else {
		message = fmt.Sprintf("Archivo '%s' creado exitosamente (vacío)", params.Path)
	}

	if params.R {
		message += " (directorios padre creados automáticamente)"
	}

	response := utils.CreateResponse(
		message,
		params,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}
