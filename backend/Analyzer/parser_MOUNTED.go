package Analyzer

import (
	"fmt"
	"strings"
)

func ParseMounted(command string) (bool, error) {

	command = strings.TrimSpace(strings.ToLower(command))

	if command != "mounted" {
		return false, fmt.Errorf("comando inválido: debe ser exactamente 'mounted'")
	}

	return true, nil
}
