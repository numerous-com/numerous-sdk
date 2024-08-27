package create

import (
	"context"
	"errors"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/token"
)

type TokenCreator interface {
	Create(ctx context.Context, input token.CreateTokenInput) (token.CreateTokenOutput, error)
}

type CreateInput struct {
	Name        string
	Description string
}

func Create(ctx context.Context, creator TokenCreator, input CreateInput) error {
	out, err := creator.Create(ctx, token.CreateTokenInput(input))

	if err == nil {
		output.PrintlnOK("Created user access token %q: %s", out.Name, out.Token)
		println("Make sure to copy your access token now. You won't be able to see it again!.")

		return nil
	}

	switch {
	case errors.Is(err, token.ErrAccessDenied):
		output.PrintError("Access denied", "Your login may have expired. Try to log out and log back in again.")
	case errors.Is(err, token.ErrUserAccessTokenAlreadyExists):
		output.PrintError("Error: %s", "", err.Error())
	case errors.Is(err, token.ErrUserAccessTokenNameInvalid):
		output.PrintError("Error: %s", "", err.Error())
	default:
		output.PrintUnknownError(err)
	}

	return err
}
