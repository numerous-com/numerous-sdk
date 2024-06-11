package gql

import (
	"errors"
	"strings"
)

var ErrAccesDenied = errors.New("access denied")

func CheckAccessDenied(err error) error {
	if strings.Contains(err.Error(), "access denied") {
		return ErrAccesDenied
	}

	return err
}
