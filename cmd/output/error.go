package output

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/dir"
)

// Prints an error message prefixed with an error symbol. Variadic arguments are
// formatted into both header and body, as if they were one string.
func PrintError(header, body string, args ...any) {
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	f := errorcross + " " + AnsiRed + header + AnsiReset + "\n" + AnsiYellow + body + AnsiReset
	fmt.Printf(f, args...)
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

func PrintErrorAppNotFound(ai appident.AppIdentifier) {
	PrintError(
		"App not found",
		"The app \"%s/%s\" cannot be found. Did you specify the correct organization and app slug?",
		ai.OrganizationSlug, ai.AppSlug,
	)
}

func PrintErrorAccessDenied(ai appident.AppIdentifier) {
	PrintError(
		"Access denied.",
		`Hint: You may have specified an organization name instead of an organization slug.
Is the organization slug %q and the app slug %q correct?`,
		ai.OrganizationSlug, ai.AppSlug,
	)
}

func PrintAppError(err error, ai appident.AppIdentifier) {
	switch {
	case errors.Is(err, app.ErrAccesDenied):
		PrintErrorAccessDenied(ai)
	case errors.Is(err, app.ErrAppNotFound):
		PrintErrorAppNotFound(ai)
	default:
		PrintErrorDetails("Error occurred for app \"%s/%s\"", err, ai.OrganizationSlug, ai.AppSlug)
	}
}

func PrintReadAppIDErrors(err error, appDir string) {
	if err == dir.ErrAppIDNotFound {
		PrintErrorAppNotInitialized(appDir)
	} else if err != nil {
		PrintErrorDetails("An error occurred reading the app ID", err)
	}
}

func PrintGetAppIdentiferError(err error, appDir string, ai appident.AppIdentifier) {
	switch {
	case errors.Is(err, appident.ErrAppNotInitialized):
		PrintErrorAppNotInitialized(appDir)
	case errors.Is(err, appident.ErrInvalidAppSlug):
		PrintErrorInvalidAppSlug(ai.AppSlug)
	case errors.Is(err, appident.ErrInvalidOrganizationSlug):
		PrintErrorInvalidOrganizationSlug(ai.OrganizationSlug)
	case errors.Is(err, appident.ErrMissingAppSlug):
		PrintErrorMissingAppSlug()
	case errors.Is(err, appident.ErrMissingOrganizationSlug):
		PrintErrorMissingOrganizationSlug()
	}
}

func PrintManifestTOMLError(err error) {
	if !strings.HasPrefix(err.Error(), "toml:") {
		return
	}

	fmt.Println("There is a an error in your \"numerous.toml\" manifest.\n" + err.Error())
}
