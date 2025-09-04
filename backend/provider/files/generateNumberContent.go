package files

import (
	"fmt"
	"strconv"
	"strings"
)

func generateNumberContent(size int32) string {
	fmt.Printf("🔢 DEBUG generateNumberContent: Generando contenido de tamaño %d\n", size)

	if size == 0 {
		return ""
	}

	var builder strings.Builder
	for i := int32(0); i < size; i++ {
		digit := i % 10
		builder.WriteString(strconv.Itoa(int(digit)))
	}

	result := builder.String()
	fmt.Printf("🔢 DEBUG generateNumberContent: Generado %d chars, primeros 50: %q\n", len(result), result[:minLen(50, len(result))])
	return result
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
