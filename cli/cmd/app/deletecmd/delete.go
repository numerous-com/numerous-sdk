package deletecmd

import (
	"context"

	"numerous/cli/cmd/app/appident"
	"numerous/cli/cmd/output"
	"numerous/cli/internal/app"
)

type AppService interface {
	Delete(ctx context.Context, input app.DeleteAppInput) error
}

func Delete(ctx context.Context, apps AppService, appDir, slug, appName string) error {
	slug, appName, err := appident.GetAppIdentifier(appDir, slug, appName)
	if err != nil {
		return err
	}

	err = apps.Delete(ctx, app.DeleteAppInput{OrganizationSlug: slug, Name: appName})
	if err != nil {
		output.PrintErrorDetails("Error occurred deleting app.", err)
		return err
	}

	output.PrintlnOK("Deleted app %s/%s", slug, appName)

	return nil
}
