package routes

import (
	"main/Analyzer"
	diskmanager "main/diskManager"
	"main/utils"
	"net/http"
)

func handlePdisk(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParsePdisk(command)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			nil,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	diskInfo, err := diskmanager.ShowDiskPartitions(params)
	if err != nil {
		response := utils.CreateResponse(
			err.Error(),
			params,
			"system-error",
		)

		utils.SendJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	formattedOutput := diskmanager.FormatPartitionsOutput(diskInfo)

	response := utils.CreateResponse(
		formattedOutput,
		diskInfo,
		"system",
	)

	utils.SendJSONResponse(w, response, http.StatusOK)
}
