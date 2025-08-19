package routes

import (
	"fmt"
	"main/Analyzer"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"net/http"
	"strings"
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

	mountManager := mounted.GetMountManager()
	partitions := mountManager.GetMountedPartitions()

	// Formatear las particiones montadas
	var formattedOutput string
	if len(partitions) == 0 {
		formattedOutput = "No hay particiones montadas actualmente en memoria"
	} else {
		var result strings.Builder
		result.WriteString("Particiones montadas en memoria:\n\n ")

		for i, partition := range partitions {
			result.WriteString(fmt.Sprintf("%d. ID: %s | Nombre: %s | Ruta: %s\n ",
				i+1, partition.ID, partition.Name, partition.Path))
		}

		formattedOutput = result.String()
	}

	response := structs.Response{
		Message: formattedOutput,
		Data:    partitions,
		Type:    "mounted-partitions",
	}

	utils.SendJSONResponse(w, response, http.StatusOK)
}
