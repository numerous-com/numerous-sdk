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

// Prints a standardized error message about the given appDir not being
// initialized.
func PrintErrorAppNotInitialized(appDir string) {
	if appDir == "." {
		PrintError("The current directory is not a numerous app",
			"Run \"numerous init\" to initialize a numerous app in the current directory.")
	} else {
		PrintError("The select directory \"%s\" is not a numerous app",
			"Run \"numerous init %s\" to initialize a numerous app.",
			appDir, appDir)
	}
}

// Print an error message requesting that the user logs in.
func PrintErrorLogin() {
	PrintError(
		"Command requires login.",
		"Use \"numerous login\" to login or sign up.\n",
	)
}

func PrintErrorMissingAppName() {
	PrintError(
		"Missing app name.",
		`An app name must be given as either a command flag, or in the "deploy" section of the app manifest.`,
	)
}

func PrintErrorMissingOrganizationSlug() {
	PrintError(
		"Missing organization identifier.",
		`An organization identifier must be given as either a command flag, or in the "deploy" section of the app manifest.`,
	)
}

func PrintErrorInvalidOrganizationSlug(slug string) {
	PrintError("Invalid organization %q.", "Must contain only lower-case alphanumerical characters and dashes.", slug)
}

func PrintErrorInvalidAppName(appName string) {
	PrintError("Invalid app name %q.", "Must contain only lower-case alphanumerical characters and dashes.", appName)
}
