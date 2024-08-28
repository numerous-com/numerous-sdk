package token

import (
	"errors"
	"strings"
)

var (
	ErrAccessDenied                     = errors.New("access denied")
	ErrPersonalAccessTokenNameInvalid   = errors.New("personal access token name invalid")
	ErrPersonalAccessTokenAlreadyExists = errors.New("personal access token already exists")
)

func ConvertErrors(err error) error {
	if strings.Contains(err.Error(), "access denied") {
		return ErrAccessDenied
	}

	return err
}
