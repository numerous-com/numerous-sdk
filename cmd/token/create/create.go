package create

import (
	"context"
	"errors"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/token"
)

var ErrMissingTokenName = errors.New("missing token name argument")

type TokenCreator interface {
	Create(ctx context.Context, input token.CreateTokenInput) (token.CreateTokenOutput, error)
}

type CreateInput struct {
	Name        string
	Description string
}

func Create(ctx context.Context, creator TokenCreator, input CreateInput) error {
	if input.Name == "" {
		output.PrintError("Missing token name argument.", "")
		return ErrMissingTokenName
	}

	out, err := creator.Create(ctx, token.CreateTokenInput(input))

	if err == nil {
		output.PrintlnOK("Created personal access token %q: %s", out.Name, out.Token)
		println("Make sure to copy your access token now. You won't be able to see it again!")

		return nil
	}

	switch {
	case errors.Is(err, token.ErrAccessDenied):
		output.PrintErrorAccessDenied()
	case errors.Is(err, token.ErrPersonalAccessTokenAlreadyExists):
		output.PrintError("Error: %s", "", err.Error())
	case errors.Is(err, token.ErrPersonalAccessTokenNameInvalid):
		output.PrintError("Error: %s", "", err.Error())
	default:
		output.PrintUnknownError(err)
	}

	return err
}
