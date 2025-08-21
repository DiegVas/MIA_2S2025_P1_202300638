package routes

import (
	"encoding/json"
	"io"
	"main/structs"
	"net/http"
	"strings"
)

func HandlePostUser(w http.ResponseWriter, r *http.Request) {

	// ! Verificaciones de entrada

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el body", http.StatusBadRequest)
		return
	}

	var command structs.Command
	err = json.Unmarshal(body, &command)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	commandLower := strings.ToLower(command.Command)

	// ? Comandos sin parámetros
	switch commandLower {
	case "mounted":
		handleMounted(w, command.Command)
		return
	case "logout":
		handleLogout(w, command.Command)
		return
	}

	commandBase := strings.Split(commandLower, " ")[0]

	// ? Comandos con parámetros

	switch commandBase {
	case "mkdisk":
		handleMkdisk(w, command.Command)
	case "rmdisk":
		handleRmdisk(w, command.Command)
	case "fdisk":
		handleFdisk(w, command.Command)
	case "mount":
		handleMount(w, command.Command)
	case "mkfs":
		handleMkfs(w, command.Command)
	case "login":
		handleLogin(w, command.Command)
	case "mkgrp":
		handleMkgrp(w, command.Command)
	case "rmgrp":
		handleRmgrp(w, command.Command)
	case "mkusr":
		handleMkusr(w, command.Command)
	case "rmusr":
		handleRmusr(w, command.Command)
	case "chgrp":
		handleChgrp(w, command.Command)
	case "cat":
		handleCat(w, command.Command)
	case "mkfile":
		handleMkfile(w, command.Command)
	case "mkdir":
		handleMkdir(w, command.Command)
	case "pdisk":
		handlePdisk(w, command.Command)
	case "viewusers":
		handleViewUsers(w, command.Command)
	case "rep":
		handleRep(w, command.Command)
	default:

		response := structs.Response{
			Message: "Comando no reconocido. Comandos soportados: MKDISK, RMDISK, FDISK, MOUNT, MOUNTED, MKFS, LOGIN, LOGOUT, MKGRP, RMGRP, MKUSR, RMUSR, CHGRP, CAT, MKFILE, MKDIR, PDISK, REP",
			Data:    nil,
			Type:    "system-error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}
}
