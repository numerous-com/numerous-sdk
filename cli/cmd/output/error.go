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

	f := errorcross + " " + ansiRed + header + ansiReset + "\n" + ansiYellow + body + ansiReset
	fmt.Printf(f, args...)
}

// Prints an error message with the given header, and a body that contains
// the error details. Variadic arguments will be used for string formatting.
func PrintErrorDetails(header string, err error, args ...any) {
	PrintError(header, "Details: "+err.Error())
}

// Prints the given error with a standardized error message.
func PrintUnknownError(err error) {
	PrintErrorDetails("Sorry! An unexpected error occurred.", err)
}

func PrintErrorAppNotInitialized() {
	PrintError("The current directory is not a numerous app",
		"Run \"numerous init\" to initialize a numerous app in the current directory.")
}
