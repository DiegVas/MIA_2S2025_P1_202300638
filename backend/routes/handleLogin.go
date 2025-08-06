package routes

import (
	"encoding/json"
	"fmt"
	"main/Analyzer"
	"main/structs"
	"net/http"
)

func handleLogin(w http.ResponseWriter, command string) {
	params, err := Analyzer.ParseLogin(command)
	if err != nil {
		response := structs.Response{
			Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
			Data:    nil,
			Type:    "system-error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	err = Analyzer.LoginUser(params)
	if err != nil {
		response := structs.Response{
			Message: err.Error(),
			Data:    nil,
			Type:    "auth-error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}
	response := structs.Response{
		Message: fmt.Sprintf("Sesión iniciada correctamente como '%s' en la partición '%s'", params.User, params.ID),
		Data:    params,
		Type:    "auth-success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleLogout(w http.ResponseWriter, command string) {
	err := Analyzer.ParseLogout(command)
	if err != nil {
		response := structs.Response{
			Message: err.Error(),
			Data:    nil,
			Type:    "system-error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	err = Analyzer.LogoutUser()
	if err != nil {
		response := structs.Response{
			Message: err.Error(),
			Data:    nil,
			Type:    "auth-error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}
	response := structs.Response{
		Message: "Sesión cerrada correctamente",
		Data:    nil,
		Type:    "auth-success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
