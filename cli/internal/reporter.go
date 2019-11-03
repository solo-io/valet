package internal

import (
	"fmt"
	"strings"
	"time"
)

func Report(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

func MapToString(values map[string]string) string {
	var entries []string
	for k, v := range values {
		entries = append(entries, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
}
