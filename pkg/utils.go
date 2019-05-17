package utils

import (
	"fmt"
	"strings"
)

func SanitizeUrl(err error, url string, errorText string) string {
	sanitizedError := strings.ReplaceAll(fmt.Sprintf("%s", err), url, "<REDACTED URL>")
	if errorText == "" {
		return fmt.Sprintf("%s", sanitizedError)
	}
	return fmt.Sprintf("%s: %s", errorText, sanitizedError)
}
