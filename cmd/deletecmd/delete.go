package deletecmd

import (
	"context"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

type appDeleter interface {
	Delete(ctx context.Context, input app.DeleteAppInput) error
}

func deleteApp(ctx context.Context, apps appDeleter, appDir, orgSlug, appSlug string) error {
	ai, err := appident.GetAppIdentifier(appDir, nil, orgSlug, appSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, appDir, ai)
		return err
	}

	if err := apps.Delete(ctx, app.DeleteAppInput(ai)); err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	output.PrintlnOK("Deleted app %s/%s", ai.OrganizationSlug, ai.AppSlug)

	return nil
}
