package users

import (
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"strconv"
	"strings"
)

func CreateUser(params *structs.UserParams) error {
	fmt.Printf("DEBUG MKUSR: Intentando crear usuario: %s en grupo: %s\n", params.User, params.Grp)

	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}
	if session.User != "root" {
		return errors.New("solo el usuario root puede crear usuarios")
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

	fmt.Printf("DEBUG MKUSR: Contenido actual de users.txt:\n%s\n", usersFile)

	lines := strings.Split(usersFile, "\n")
	maxUID := 1
	groupExists := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) >= 3 && fields[1] == "G" && fields[0] != "0" && fields[2] == params.Grp {
			groupExists = true
			fmt.Printf("DEBUG MKUSR: Grupo '%s' encontrado con GID %s\n", params.Grp, fields[0])
		}
		if len(fields) == 5 && fields[1] == "U" {
			// Verificar si el usuario ya existe (fields[3] es el nombre de usuario)
			fmt.Printf("DEBUG MKUSR: Usuario encontrado: UID=%s, Usuario='%s'\n", fields[0], fields[3])
			if fields[3] == params.User && fields[0] != "0" {
				fmt.Printf("DEBUG MKUSR: El usuario '%s' ya existe con UID %s\n", params.User, fields[0])
				return fmt.Errorf("el usuario '%s' ya existe", params.User)
			}
			uid, _ := strconv.Atoi(fields[0])
			if uid > maxUID {
				maxUID = uid
			}
		}
	}
	if !groupExists {
		return fmt.Errorf("el grupo '%s' no existe", params.Grp)
	}

	// Crear el nuevo usuario con el siguiente UID
	newUID := maxUID + 1
	newUserLine := fmt.Sprintf("%d, U, %s, %s, %s", newUID, params.Grp, params.User, params.Pass)

	fmt.Printf("DEBUG MKUSR: Creando nuevo usuario - UID: %d, Línea: '%s'\n", newUID, newUserLine)
	if !strings.HasSuffix(usersFile, "\n") {
		usersFile += "\n"
	}
	usersFile += newUserLine + "\n"

	err = WriteUsersFile(partition.Path, partition.Name, usersFile)
	if err != nil {
		fmt.Printf("ERROR en WriteUsersFile: %v\n", err)
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}

	return nil
}
