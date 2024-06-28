package deletecmd

import (
	"context"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type AppService interface {
	Delete(ctx context.Context, input app.DeleteAppInput) error
}

func Delete(ctx context.Context, apps AppService, appDir, orgSlug, appSlug string) error {
	ai, err := appident.GetAppIdentifier(appDir, nil, orgSlug, appSlug)
	if err != nil {
		return err
	}

	err = apps.Delete(ctx, app.DeleteAppInput(ai))
	if err != nil {
		output.PrintErrorDetails("Error occurred deleting app.", err)
		return err
	}

	output.PrintlnOK("Deleted app %s/%s", ai.OrganizationSlug, ai.AppSlug)

	return nil
}
