package token

import (
	"errors"
	"strings"
)

var (
	ErrAccessDenied                 = errors.New("access denied")
	ErrUserAccessTokenNameInvalid   = errors.New("user access token name invalid")
	ErrUserAccessTokenAlreadyExists = errors.New("user access token already exists")
)

func ConvertErrors(err error) error {
	if strings.Contains(err.Error(), "access denied") {
		return ErrAccessDenied
	}

	return err
}
