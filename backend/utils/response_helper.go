package utils

import "main/structs"

func GetMessageType(messageType string) string {
	if messageType == "system-error" || messageType == "auth-error" {
		return messageType
	}

	session := structs.GetSession()
	if !session.Active {
		if messageType == "system" || messageType == "user" {
			return "system"
		}
		return messageType
	}

	if session.User == "root" {
		return "root"
	} else {
		return "default"
	}
}

func CreateResponse(message string, data interface{}, baseType string) structs.Response {
	return structs.Response{
		Message: message,
		Data:    data,
		Type:    GetMessageType(baseType),
	}
}
