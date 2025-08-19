package files

import (
	"bytes"
	"encoding/binary"
	"main/provider/mounted"
	"main/structs"
	"os"
)

func HasWritePermissionOnDirectory(file *os.File, dirInodeNum int32, session *structs.Session, inodePos int32) bool {
	// ? Leer el inodo del directorio
	inodeBytes := make([]byte, structs.INODE_SIZE)
	_, err := file.ReadAt(inodeBytes, int64(inodePos+(dirInodeNum*structs.INODE_SIZE)))
	if err != nil {
		return false
	}

	var inode structs.Inode
	buffer := bytes.NewReader(inodeBytes)
	binary.Read(buffer, binary.LittleEndian, &inode)

	// ? Si es root, tiene todos los permisos
	if session.User == "root" {
		return true
	}

	// ? Obtener información del usuario actual
	currentUser, err := GetCurrentUserInfo(session, mounted.MountedPartition{})
	if err != nil {
		return false
	}

	// ? Verificar permisos de escritura
	perms := inode.I_perm

	// ? Verificar permisos de propietario
	if inode.I_uid == currentUser.UID {
		return (perms & 0200) != 0 // Permiso de escritura del propietario
	}

	// ? Verificar permisos de grupo
	if inode.I_gid == currentUser.GID {
		return (perms & 0020) != 0 // Permiso de escritura del grupo
	}

	// ? Verificar permisos de otros
	return (perms & 0002) != 0 // Permiso de escritura de otros
}
