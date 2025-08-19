package files

import (
	"main/structs"
)

func hasReadPermission(inode structs.Inode, user *structs.User) bool {
	if user == nil {
		return false
	}

	// Si es root, tiene todos los permisos
	if user.UID == 1 {
		return true
	}

	// Obtener los permisos del archivo
	perms := inode.I_perm

	// Verificar permisos de propietario
	if inode.I_uid == user.UID {
		return (perms & 0400) != 0 // Permiso de lectura del propietario
	}

	// Verificar permisos de grupo
	if inode.I_gid == user.GID {
		return (perms & 0040) != 0 // Permiso de lectura del grupo
	}

	// Verificar permisos de otros
	return (perms & 0004) != 0 // Permiso de lectura de otros
}
