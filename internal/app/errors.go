package app

import (
	"errors"
	"strings"
)

var (
	ErrAppNotFound  = errors.New("app not found")
	ErrAccessDenied = errors.New("access denied")
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
