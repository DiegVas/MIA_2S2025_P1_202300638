package groups

import (
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/provider/users"
	"main/structs"
	"strconv"
	"strings"
)

func RemoveGroup(params *structs.GroupParams) error {
	// Verificar la sesión activa
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}

	if session.User != "root" {
		return errors.New("solo el usuario root puede eliminar grupos")
	}

	// Obtener la partición montada de la sesión actual
	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}

	// Leer el contenido actual de users.txt
	usersFile, err := users.ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	// Buscar y marcar el grupo como eliminado
	lines := strings.Split(usersFile, "\n")
	groupFound := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		for j := range fields {
			fields[j] = strings.TrimSpace(fields[j])
		}

		// Si es una línea de grupo (G) y coincide el nombre
		if len(fields) >= 3 && fields[1] == "G" && fields[2] == params.Name {
			gid, _ := strconv.Atoi(fields[0])
			if gid > 0 {
				// Verificar que no sea el grupo root
				if params.Name == "root" {
					return errors.New("no se puede eliminar el grupo root")
				}

				// Marcar como eliminado (GID = 0)
				lines[i] = fmt.Sprintf("0, G, %s", params.Name)
				groupFound = true
				break
			}
		}
	}

	if !groupFound {
		return fmt.Errorf("el grupo '%s' no existe o ya fue eliminado", params.Name)
	}

	// Reconstruir el archivo
	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	// Escribir el archivo actualizado
	err = users.WriteUsersFile(partition.Path, partition.Name, newContent)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}

	return nil
}
