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

func CreateGroup(params *structs.GroupParams) error {

	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay una sesión activa, debe iniciar sesión primero")
	}

	if session.User != "root" {
		return errors.New("solo el usuario root puede crear grupos")
	}

	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(session.ID)
	if !found {
		return fmt.Errorf("no se encontró la partición con ID '%s'", session.ID)
	}

	usersFile, err := users.ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	fmt.Printf("DEBUG MKGRP: Contenido actual de users.txt:\n%s\n", usersFile)
	fmt.Printf("DEBUG MKGRP: Intentando crear grupo: %s\n", params.Name)

	lines := strings.Split(usersFile, "\n")
	maxGID := 1 // GID mínimo (root es 1)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		// Si es una línea de grupo (G)
		if len(fields) >= 3 && fields[1] == "G" {
			// Verificar si el grupo existe y no está eliminado (GID > 0)
			gid, _ := strconv.Atoi(fields[0])
			fmt.Printf("DEBUG MKGRP: Encontrado grupo existente: GID=%d, nombre='%s'\n", gid, fields[2])

			if gid > 0 && fields[2] == params.Name {
				fmt.Printf("DEBUG MKGRP: El grupo '%s' ya existe con GID %d\n", params.Name, gid)
				return fmt.Errorf("el grupo '%s' ya existe", params.Name)
			}
			if gid > maxGID {
				maxGID = gid
			}
		}
	}

	// Crear el nuevo grupo con el siguiente GID
	newGID := maxGID + 1
	newGroupLine := fmt.Sprintf("%d, G, %s", newGID, params.Name)

	fmt.Printf("DEBUG MKGRP: Creando nuevo grupo - GID: %d, Línea: '%s'\n", newGID, newGroupLine)

	// Agregar la nueva línea al archivo
	if !strings.HasSuffix(usersFile, "\n") {
		usersFile += "\n"
	}
	usersFile += newGroupLine + "\n"

	fmt.Printf("DEBUG MKGRP: Contenido actualizado de users.txt:\n%s\n", usersFile)

	// Escribir el archivo actualizado
	err = users.WriteUsersFile(partition.Path, partition.Name, usersFile)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo de usuarios: %v", err)
	}

	return nil
}
