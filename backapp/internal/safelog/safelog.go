package safelog

import (
	"fmt"
	"strings"
)

// Value converts an untrusted value into a single-line string suitable for logs.
func Value(value any) string {
	result := fmt.Sprint(value)
	result = strings.ReplaceAll(result, "\r", `\r`)
	result = strings.ReplaceAll(result, "\n", `\n`)
	return result
}
