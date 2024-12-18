package set

import (
	"errors"

	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/output"
)

var (
	ErrInvalidPropertyToSet = errors.New("invalid property to set")
	ErrInvalidArgumentCount = errors.New("expected exactly two arguments")
)

func configSet(args []string) error {
	if len(args) != 2 { // nolint:mnd
		output.PrintError("Expected exactly two arguments, got %d", "", len(args))
		return ErrInvalidArgumentCount
	}
	property, value := args[0], args[1]

	cfg := config.Config{}
	if err := cfg.Load(); err != nil {
		output.PrintErrorDetails("Error loading configuration", err)
		return err
	}

	switch property {
	case "organization":
		cfg.OrganizationSlug = value
	default:
		output.PrintError("Invalid property to set: %q", "", property)
		return ErrInvalidPropertyToSet
	}

	if err := cfg.Save(); err != nil {
		output.PrintErrorDetails("Error saving configuration", err)
		return err
	}

	output.PrintlnOK("Set %s=%q", property, value)

	return nil
}
