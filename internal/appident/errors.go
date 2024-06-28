package appident

import "errors"

var (
	ErrInvalidOrganizationSlug = errors.New("invalid organization slug")
	ErrInvalidAppSlug          = errors.New("invalid app slug")
)
