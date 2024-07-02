package app

import (
	"errors"
	"strings"
)

var (
	ErrAppNotFound = errors.New("app not found")
	ErrAccesDenied = errors.New("access denied")
)

func ConvertErrors(err error) error {
	if strings.Contains(err.Error(), "access denied") {
		return ErrAccesDenied
	}

	if strings.Contains(err.Error(), "app not found") {
		return ErrAppNotFound
	}

	return err
}
