package users

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"strings"
)

type UserInfo struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Group    string `json:"group"`
	Password string `json:"password"`
	Status   string `json:"status"`
}

type GroupInfo struct {
	GID    string `json:"gid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ViewUsersResponse struct {
	Users  []UserInfo  `json:"users"`
	Groups []GroupInfo `json:"groups"`
}

func ViewAllUsers(params *structs.ViewUsersParams) (*ViewUsersResponse, error) {
	// Verificar que hay una sesión activa
	session := structs.GetSession()
	if !session.Active {
		return nil, fmt.Errorf("no hay una sesión activa, debe iniciar sesión primero")
	}

	// Solo el usuario root puede ejecutar este comando
	if session.User != "root" {
		return nil, fmt.Errorf("solo el usuario root puede ver la información de usuarios y grupos")
	}

	// Obtener la partición montada
	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(params.ID)
	if !found {
		return nil, fmt.Errorf("no se encontró la partición con ID '%s'", params.ID)
	}

	// Leer el archivo de usuarios
	usersFile, err := ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	response := &ViewUsersResponse{
		Users:  []UserInfo{},
		Groups: []GroupInfo{},
	}

	lines := strings.Split(usersFile, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		if len(fields) < 3 {
			continue
		}

		uid := fields[0]
		entryType := fields[1]

		// Procesar grupos (formato: uid, G, nombre)
		if entryType == "G" && len(fields) >= 3 {
			status := "activo"
			if uid == "0" {
				status = "eliminado"
			}

			groupInfo := GroupInfo{
				GID:    uid,
				Name:   fields[2],
				Status: status,
			}
			response.Groups = append(response.Groups, groupInfo)
		}

		// Procesar usuarios (formato: uid, U, grupo, usuario, password)
		if entryType == "U" && len(fields) >= 5 {
			status := "activo"
			if uid == "0" {
				status = "eliminado"
			}

			userInfo := UserInfo{
				UID:      uid,
				Username: fields[3],
				Group:    fields[2],
				Password: fields[4],
				Status:   status,
			}
			response.Users = append(response.Users, userInfo)
		}
	}

	return response, nil
}

// GenerateDetailedMessage crea un mensaje detallado con toda la información de usuarios y grupos
func GenerateDetailedMessage(response *ViewUsersResponse) string {
	var message strings.Builder

	message.WriteString("=== INFORMACIÓN DE USUARIOS Y GRUPOS ===\n")
	message.WriteString(fmt.Sprintf("Total de usuarios: %d\n", len(response.Users)))
	message.WriteString(fmt.Sprintf("Total de grupos: %d\n\n", len(response.Groups)))

	// Mostrar información de grupos
	if len(response.Groups) > 0 {
		message.WriteString("--- GRUPOS ---\n")
		for _, group := range response.Groups {
			message.WriteString(fmt.Sprintf("GID: %s | Nombre: %s | Estado: %s\n",
				group.GID, group.Name, group.Status))
		}
		message.WriteString("\n")
	}

	// Mostrar información de usuarios
	if len(response.Users) > 0 {
		message.WriteString("--- USUARIOS ---\n")
		for _, user := range response.Users {
			message.WriteString(fmt.Sprintf("UID: %s | Usuario: %s | Grupo: %s | Contraseña: %s | Estado: %s\n",
				user.UID, user.Username, user.Group, user.Password, user.Status))
		}
	}

	return message.String()
}
