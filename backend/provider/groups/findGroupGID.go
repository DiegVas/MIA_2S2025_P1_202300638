package groups

import (
	"strconv"
	"strings"
)

func FindGroupGID(usersContent, groupName string) int {
	lines := strings.Split(usersContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 && strings.TrimSpace(parts[1]) == "G" {
			if strings.TrimSpace(parts[2]) == groupName {
				gid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				return gid
			}
		}
	}
	return 0
}
