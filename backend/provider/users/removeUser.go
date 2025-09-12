package users

import (
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"strings"
)

func RemoveUser(params *structs.RmUserParams) error {
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	if session.User != "root" {
		return errors.New("solo el usuario root puede eliminar usuarios")
	}
	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}
	usersFile, err := ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}
	lines := strings.Split(usersFile, "\n")
	userFound := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for j := range fields {
			fields[j] = strings.TrimSpace(fields[j])
		}
		if len(fields) == 5 && fields[1] == "U" && fields[3] == params.User && fields[0] != "0" {
			lines[i] = fmt.Sprintf("0, U, %s, %s, %s", fields[2], fields[3], fields[4])
			userFound = true
			break
		}
	}
	if !userFound {
		return fmt.Errorf("el usuario '%s' no existe o ya fue eliminado", params.User)
	}
	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	err = WriteUsersFile(partition.Path, partition.Name, newContent)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}
	return nil
}
