package app

import (
	"errors"
	"strings"

	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

var (
	ErrAppNotFound          = errors.New("app not found")
	ErrAccessDenied         = errors.New("access denied")
	ErrDeploymentNotFound   = errors.New("deployment not found")
	ErrOrganizationNotFound = errors.New("organization not found")
)

func convertErrors(err error) error {
	if strings.Contains(err.Error(), "access denied") {
		return ErrAccessDenied
	}

	if strings.Contains(err.Error(), "app not found") {
		return ErrAppNotFound
	}

	return err
}

func PrintAppError(err error, ai appident.AppIdentifier) {
	switch {
	case errors.Is(err, ErrAccessDenied):
		PrintErrorAccessDenied(ai)
	case errors.Is(err, ErrAppNotFound):
		PrintErrorAppNotFound(ai)
	default:
		output.PrintErrorDetails("Error occurred for app \"%s/%s\"", err, ai.OrganizationSlug, ai.AppSlug)
	}
}

func PrintErrorAppNotFound(ai appident.AppIdentifier) {
	output.PrintError(
		"App not found",
		"The app \"%s/%s\" cannot be found. Did you specify the correct organization and app slug?",
		ai.OrganizationSlug, ai.AppSlug,
	)
}

func PrintErrorAccessDenied(ai appident.AppIdentifier) {
	output.PrintError(
		"Access denied.",
		`Hint: You may have specified an organization name instead of an organization slug.
Is the organization slug %q and the app slug %q correct?`,
		ai.OrganizationSlug, ai.AppSlug,
	)
}

func PrintErrorOrganizationNotFound(ai appident.AppIdentifier) {
	output.PrintError(
		"Organization not found",
		"The organization \"%s\" cannot be found. Did you specify the correct organization slug?",
		ai.OrganizationSlug,
	)
}

func PrintTaskError(err error, ai appident.AppIdentifier) {
	switch {
	case errors.Is(err, ErrAccessDenied):
		PrintErrorAccessDenied(ai)
	case errors.Is(err, ErrAppNotFound):
		PrintErrorAppNotFound(ai)
	case errors.Is(err, ErrOrganizationNotFound):
		PrintErrorOrganizationNotFound(ai)
	case err != nil:
		output.PrintUnknownError(err)
	}
}
