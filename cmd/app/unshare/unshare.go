package unshare

import (
	"context"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

type Input struct {
	AppDir  string
	AppSlug string
	OrgSlug string
}

type AppService interface {
	UnshareApp(ctx context.Context, ai appident.AppIdentifier) error
}

func unshareApp(ctx context.Context, apps AppService, input Input) error {
	ai, err := appident.GetAppIdentifier(input.AppDir, nil, input.OrgSlug, input.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, input.AppDir, ai)
		return err
	}

	err = apps.UnshareApp(ctx, ai)
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	output.PrintlnOK("Shared URL for %q has been removed", ai.String())

	return nil
}
