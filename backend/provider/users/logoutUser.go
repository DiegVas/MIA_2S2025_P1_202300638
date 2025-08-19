package users

import (
	"errors"
	"main/structs"
)

func LogoutUser() error {
	session := structs.GetSession()
	if !session.Active {
		return errors.New("no hay sesión activa para cerrar")
	}
	session.Logout()
	return nil
}
