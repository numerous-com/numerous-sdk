package appident

import "errors"

var (
	ErrMissingOrganizationSlug = errors.New("missing organization slug")
	ErrInvalidOrganizationSlug = errors.New("invalid organization slug")
	ErrInvalidAppSlug          = errors.New("invalid app slug")
	ErrMissingAppSlug          = errors.New("missing app slug")
	ErrAppNotInitialized       = errors.New("app not initialized")
)
