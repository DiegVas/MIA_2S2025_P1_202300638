package utils

import (
	"encoding/json"
	"main/structs"
	"net/http"
)

func SendJSONResponse(w http.ResponseWriter, response structs.Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
