package errorhandling

import "errors"

// Used to signal that an error occurred, but it should not be printed
var ErrAlreadyPrinted = errors.New("represents an error that was printed previously")

func ErrorAlreadyPrinted(err error) error {
	if err == nil {
		return nil
	} else {
		return ErrAlreadyPrinted
	}
}
