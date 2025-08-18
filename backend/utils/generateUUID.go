package utils

import "crypto/rand"

func GenerateUUID() []byte {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return uuid
}
