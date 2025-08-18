package routes

import (
	"fmt"
	"main/Analyzer"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleMounted(w http.ResponseWriter, command string) {
	valid, err := Analyzer.ParseMounted(command)

	if err != nil {
		fmt.Printf("handleMounted - error: %s\n", err.Error())

		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	if !valid {
		response := structs.Response{
			Message: "Comando inválido",
			Data:    nil,
			Type:    "system-error",
		}

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	partitions := Analyzer.GetMountedPartitions()

	formattedOutput := Analyzer.FormatMountedPartitions()

	response := structs.Response{
		Message: formattedOutput,
		Data:    partitions,
		Type:    "mounted-partitions",
	}

	utils.SendJSONResponse(w, response, http.StatusOK)
}
