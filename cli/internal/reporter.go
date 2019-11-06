package internal

import (
	"fmt"
	"time"
)

func Report(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}
