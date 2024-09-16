package appident

import (
	"errors"

	"numerous.com/cli/cmd/output"
)

var (
	ErrMissingOrganizationSlug = errors.New("missing organization slug")
	ErrInvalidOrganizationSlug = errors.New("invalid organization slug")
	ErrInvalidAppSlug          = errors.New("invalid app slug")
	ErrMissingAppSlug          = errors.New("missing app slug")
	ErrAppNotInitialized       = errors.New("app not initialized")
)

func PrintGetAppIdentiferError(err error, appDir string, ai AppIdentifier) {
	switch {
	case errors.Is(err, ErrAppNotInitialized):
		output.PrintErrorAppNotInitialized(appDir)
	case errors.Is(err, ErrInvalidAppSlug):
		output.PrintErrorInvalidAppSlug(ai.AppSlug)
	case errors.Is(err, ErrInvalidOrganizationSlug):
		output.PrintErrorInvalidOrganizationSlug(ai.OrganizationSlug)
	case errors.Is(err, ErrMissingAppSlug):
		output.PrintErrorMissingAppSlug()
	case errors.Is(err, ErrMissingOrganizationSlug):
		output.PrintErrorMissingOrganizationSlug()
	}
}
