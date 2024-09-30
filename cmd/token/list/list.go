package list

import (
	"context"
	"errors"
	"time"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/token"
)

type TokenLister interface {
	List(ctx context.Context) (token.ListTokenOutput, error)
}

func List(ctx context.Context, lister TokenLister) error {
	out, err := lister.List(ctx)

	switch {
	case err == nil:
		list(out)
	case errors.Is(err, token.ErrAccessDenied):
		output.PrintErrorAccessDenied()
	default:
		output.PrintUnknownError(err)
	}

	return err
}

func list(tokens token.ListTokenOutput) {
	for i, token := range tokens {
		if i > 0 {
			println()
		}

		printToken(token)
	}
}

func printToken(token token.TokenEntry) {
	expirationValue := "never"
	if token.ExpiresAt != nil {
		expirationValue = token.ExpiresAt.Format(time.RFC3339)
	}

	println("Name:        " + token.Name)
	println("UUID:        " + token.ID)
	println("Description: " + token.Description)
	println("Expires at:  " + expirationValue)
}
