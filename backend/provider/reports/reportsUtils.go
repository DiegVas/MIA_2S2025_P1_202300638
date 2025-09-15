package reports

import (
	"fmt"
	"strings"
	"time"
)

// formatInodeTime formatea el tiempo del inodo
func formatInodeTime(timeBytes [16]byte) string {
	timeStr := strings.TrimRight(string(timeBytes[:]), "\x00")
	if timeStr == "" {
		return "N/A"
	}

	// Intentar parsear diferentes formatos de fecha
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}

	return timeStr
}

// formatPermissions convierte los permisos numéricos a formato rwx
func formatPermissions(perm int32) string {
	if perm == 0 {
		return "----------"
	}

	// Determinar tipo de archivo (primer carácter)
	fileType := "-"

	// Los siguientes 9 caracteres son rwx para owner, group, others
	perms := make([]byte, 9)

	// Owner permissions (bits 8-6)
	if perm&0400 != 0 {
		perms[0] = 'r'
	} else {
		perms[0] = '-'
	}
	if perm&0200 != 0 {
		perms[1] = 'w'
	} else {
		perms[1] = '-'
	}
	if perm&0100 != 0 {
		perms[2] = 'x'
	} else {
		perms[2] = '-'
	}

	// Group permissions (bits 5-3)
	if perm&0040 != 0 {
		perms[3] = 'r'
	} else {
		perms[3] = '-'
	}
	if perm&0020 != 0 {
		perms[4] = 'w'
	} else {
		perms[4] = '-'
	}
	if perm&0010 != 0 {
		perms[5] = 'x'
	} else {
		perms[5] = '-'
	}

	// Others permissions (bits 2-0)
	if perm&0004 != 0 {
		perms[6] = 'r'
	} else {
		perms[6] = '-'
	}
	if perm&0002 != 0 {
		perms[7] = 'w'
	} else {
		perms[7] = '-'
	}
	if perm&0001 != 0 {
		perms[8] = 'x'
	} else {
		perms[8] = '-'
	}

	return fileType + string(perms)
}

// isValidFilename verifica si un nombre de archivo es válido
func isValidFilename(name string) bool {
	if name == "" || len(name) > 12 {
		return false
	}

	// Verificar que solo contenga caracteres imprimibles válidos
	for _, r := range name {
		if r < 32 || r > 126 {
			return false
		}
	}

	// Verificar que no contenga caracteres problemáticos
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, invalid := range invalidChars {
		if strings.Contains(name, invalid) {
			return false
		}
	}

	return true
}

// escapeHTMLForDot escapa caracteres especiales para uso en HTML dentro de archivos DOT
func escapeHTMLForDot(s string) string {
	original := s
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")

	// Escapar caracteres nulos y otros caracteres no imprimibles
	var result strings.Builder
	for _, char := range s {
		if char == 0 {
			result.WriteString("\\0")
		} else if char < 32 && char != 9 && char != 10 && char != 13 { // Escapar caracteres de control excepto tab, LF, CR
			result.WriteString(fmt.Sprintf("\\x%02x", char))
		} else {
			result.WriteRune(char)
		}
	}
	s = result.String()

	if original != s {

	}

	return s
}
