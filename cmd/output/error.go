package output

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Prints an error message prefixed with an error symbol. Variadic arguments are
// formatted into both header and body, as if they were one string.
func PrintError(header, body string, args ...any) {
	body = addTrailingNewLine(body)
	f := errorcross + " " + AnsiRed + header + AnsiReset + "\n" + AnsiYellow + body + AnsiReset
	fmt.Printf(f, args...)
}

func PrintWarning(header, body string) {
	body = addTrailingNewLine(body)
	f := AnsiYellow + "! " + header + "\n" + body + AnsiReset
	fmt.Print(f)
}

func addTrailingNewLine(value string) string {
	if value != "" && !strings.HasSuffix(value, "\n") {
		value += "\n"
	}

	return value
}

// Prints an error message with the given header, and a body that contains
// the error details. Variadic arguments will be used for string formatting.
func PrintErrorDetails(header string, err error, args ...any) {
	PrintError(header, "Details: "+err.Error(), args...)
}

// Prints the given error with a standardized error message.
func PrintUnknownError(err error) {
	PrintErrorDetails("Sorry! An unexpected error occurred.", err)
}

// Prints a standardized error message about the given appDir not being
// initialized.
func PrintErrorAppNotInitialized(appDir string) {
	if appDir == "." || appDir == "" {
		PrintError("The current directory is not a numerous app",
			"Run \"numerous init\" to initialize a numerous app in the current directory.")
	} else {
		PrintError("The selected directory \"%s\" is not a numerous app",
			"Run \"numerous init %s\" to initialize a numerous app.",
			appDir, appDir)
	}
}

// Prints an error message requesting that the user logs in.
func PrintErrorLogin() {
	PrintError(
		"Command requires login.",
		"Use \"numerous login\" to login or sign up.\n",
	)
}

func PrintErrorLoginForCommand(cmd *cobra.Command) {
	fmt.Println("The command " + Highlight(cmd.CommandPath()) + " can only be used when logged in." + symbol(" "+raiseHand))
	fmt.Println("Use " + Highlight("numerous login") + " to enable this command." + symbol(" "+shootingStar))
}

func PrintErrorMissingAppSlug() {
	PrintError(
		"Missing app slug.",
		`An app slug must be given as either a command flag, or in the "deploy" section of the app manifest.`,
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

func PrintErrorInvalidAppSlug(appSlug string) {
	PrintError("Invalid app %q.", "Must contain only lower-case alphanumerical characters and dashes.", appSlug)
}

func PrintManifestTOMLError(err error) {
	if !strings.HasPrefix(err.Error(), "toml:") {
		return
	}

	fmt.Println("There is a an error in your \"numerous.toml\" manifest.\n" + err.Error())
}

func PrintErrorAccessDenied() {
	PrintError("Access denied", "Your login may have expired. Try to log out and log back in again.")
}
