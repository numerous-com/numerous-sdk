package app

import (
	"encoding/base64"
	"unicode/utf8"
)

func DecodeTaskInputForDisplay(base64Input *string) string {
	if base64Input == nil {
		return "(none)"
	}

	decoded, err := base64.StdEncoding.DecodeString(*base64Input)
	if err != nil {
		return "(base64) " + *base64Input
	}

	if !utf8.Valid(decoded) {
		return "(binary data)"
	}

	return string(decoded)
}

func TruncateInputForDisplay(input string, maxLen int) string {
	if input == "(none)" || input == "(binary data)" {
		return input
	}

	if len(input) <= maxLen {
		return input
	}

	return input[:maxLen-3] + "..."
}
