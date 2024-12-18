package share

import (
	"context"
	"errors"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

var ErrEmptySharedURL = errors.New("empty shared URL")

type Input struct {
	AppDir  string
	AppSlug string
	OrgSlug string
}

type AppService interface {
	ShareApp(ctx context.Context, ai appident.AppIdentifier) (app.ShareAppOutput, error)
}

func shareApp(ctx context.Context, apps AppService, input Input) error {
	ai, err := appident.GetAppIdentifier(input.AppDir, nil, input.OrgSlug, input.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, input.AppDir, ai)
		return err
	}

	out, err := apps.ShareApp(ctx, ai)
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	} else if out.SharedURL == nil {
		return ErrEmptySharedURL
	}

	output.PrintlnOK("Shared URL for %q:\n\t%s", ai.String(), *out.SharedURL)

	return nil
}
