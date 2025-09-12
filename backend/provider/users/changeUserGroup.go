package users

import (
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"strings"
)

func ChangeUserGroup(params *structs.ChgrpParams) error {
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	if session.User != "root" {
		return errors.New("solo el usuario root puede cambiar grupos de usuarios")
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

	groupExists := false
	userFound := false
	userLineIndex := -1

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for j := range fields {
			fields[j] = strings.TrimSpace(fields[j])
		}

		if len(fields) >= 3 && fields[1] == "G" && fields[0] != "0" && fields[2] == params.Grp {
			groupExists = true
		}

		if len(fields) == 5 && fields[1] == "U" && fields[3] == params.User && fields[0] != "0" {
			userFound = true
			userLineIndex = i
		}
	}

	if !groupExists {
		return fmt.Errorf("el grupo '%s' no existe o está eliminado", params.Grp)
	}

	if !userFound {
		return fmt.Errorf("el usuario '%s' no existe o ya fue eliminado", params.User)
	}

	line := strings.TrimSpace(lines[userLineIndex])
	fields := strings.Split(line, ",")
	for j := range fields {
		fields[j] = strings.TrimSpace(fields[j])
	}

	fields[2] = params.Grp
	lines[userLineIndex] = fmt.Sprintf("%s, %s, %s, %s, %s", fields[0], fields[1], fields[2], fields[3], fields[4])

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
