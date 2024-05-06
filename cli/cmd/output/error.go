package output

import (
	"fmt"
	"strings"
)

// Prints an error message prefixed with an error symbol. Variadic arguments are
// formatted into both header and body, as if they were one string.
func PrintError(header, body string, args ...any) {
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	f := "❌ " + header + "\n" + body
	fmt.Printf(f, args...)
}

// Prints an error message with the given header, and a body that contains
// the error details. Variadic arguments will be used for string formatting.
func PrintErrorDetails(header string, err error, args ...any) {
	PrintError(header, "Error details: "+err.Error())
}

// Prints the given error with a standardized error message.
func PrintUnknownError(err error) {
	PrintError("Sorry! An unexpected error occurred: %q", err.Error())
}
