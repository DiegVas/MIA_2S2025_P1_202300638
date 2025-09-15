package files

import (
	"strconv"
	"strings"
)

func generateNumberContent(size int32) string {

	if size == 0 {
		return ""
	}

	var builder strings.Builder
	for i := int32(0); i < size; i++ {
		digit := i % 10
		builder.WriteString(strconv.Itoa(int(digit)))
	}

	result := builder.String()
	return result
}
