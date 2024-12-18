package revoke

import (
	"context"
	"errors"

	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/token"
)

var ErrMissingTokenID = errors.New("missing token id argument")

type TokenRevoker interface {
	Revoke(ctx context.Context, id string) (token.RevokeTokenOutput, error)
}

func Revoke(ctx context.Context, revoker TokenRevoker, id string) error {
	if id == "" {
		output.PrintError("Missing token id argument.", "")
		return ErrMissingTokenID
	}

	out, err := revoker.Revoke(ctx, id)

	switch {
	case err == nil:
		output.PrintlnOK("Revoked personal access token %q", out.Name)
	case errors.Is(err, token.ErrAccessDenied):
		output.PrintErrorAccessDenied()
	default:
		output.PrintUnknownError(err)
	}

	return err
}
