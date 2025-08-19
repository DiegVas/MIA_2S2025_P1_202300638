package files

import (
	"fmt"
	"main/provider/groups"
	"main/provider/mounted"
	"main/provider/users"
	"main/structs"
	"strconv"
	"strings"
)

func GetCurrentUserInfo(session *structs.Session, partition mounted.MountedPartition) (*structs.User, error) {
	usersContent, err := users.ReadUsersFile(partition.Path, partition.Name)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de usuarios: %v", err)
	}

	lines := strings.Split(usersContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		if len(parts) >= 5 && strings.TrimSpace(parts[1]) == "U" {
			username := strings.TrimSpace(parts[3])
			if username == session.User {

				uid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				groupName := strings.TrimSpace(parts[2])

				gid := groups.FindGroupGID(usersContent, groupName)

				return &structs.User{
					Username: username,
					Password: strings.TrimSpace(parts[4]),
					Group:    groupName,
					UID:      int32(uid),
					GID:      int32(gid),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("usuario '%s' no encontrado", session.User)
}
