package revoke

import (
	"context"
	"errors"

	"numerous.com/cli/cmd/output"
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

	if err == nil {
		output.PrintlnOK("Revoked personal access token %q", out.Name)
	} else {
		output.PrintUnknownError(err)
	}

	return err
}
