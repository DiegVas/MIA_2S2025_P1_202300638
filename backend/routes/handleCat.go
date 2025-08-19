package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/files"
	"main/utils"
	"net/http"
	"strings"
)

func handleCat(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseCat(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	var results []string
	var hasErrors bool

	for _, filePath := range params.Files {
		content, err := files.ReadFileContent(filePath)
		if err != nil {
			results = append(results, fmt.Sprintf("Error al leer '%s': %s", filePath, err.Error()))
			hasErrors = true
		} else {
			results = append(results, content)
		}
	}

	finalContent := strings.Join(results, "\n")

	var messageType string
	if hasErrors {
		messageType = "system-error"
	} else {
		messageType = "system"
	}

	response := utils.CreateResponse(
		finalContent,
		params,
		messageType,
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}
