package users

import (
	"bufio"
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"strings"
)

func LoginUser(params *structs.LoginParams) error {
	session := structs.GetSession()

	if session.Active {
		return errors.New("ya hay una sesión activa, debe hacer logout primero")
	}

	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(params.ID)

	if !found {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s'", params.ID)
	}

	// Leer el contenido del archivo users.txt usando la función correcta
	content, err := ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return fmt.Errorf("error al leer users.txt: %v", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) == 5 && fields[1] == "U" {
			if fields[3] == params.User {
				if fields[4] == params.Pass {
					session.Login(fields[3], params.ID, fields[2])
					return nil
				} else {
					return errors.New("autenticación fallida: contraseña incorrecta")
				}
			}
		}
	}
	return errors.New("usuario no existe")
}
