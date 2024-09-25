package list

import (
	"context"
	"errors"
	"time"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
)

type AppLister interface {
	List(ctx context.Context, organizationSlug string) ([]app.ListApp, error)
}

type AppListInput struct {
	OrganizationSlug string
}

func list(ctx context.Context, lister AppLister, params AppListInput) error {
	apps, err := lister.List(ctx, params.OrganizationSlug)
	if err != nil {
		printAppListError(err)
		return err
	}

	emptyLineBefore := false
	for _, app := range apps {
		if emptyLineBefore {
			println()
		} else {
			emptyLineBefore = true
		}
		printApp(app)
	}

	return nil
}

func printAppListError(err error) {
	switch {
	case errors.Is(err, app.ErrAccessDenied):
		output.PrintError("Access denied", "")
	case errors.Is(err, app.ErrOrganizationNotFound):
		output.PrintError("Organization not found", "")
	case err != nil:
		output.PrintErrorDetails("Sorry! An unexpected error occurred listing apps", err)
	}
}

func printApp(app app.ListApp) {
	println("Name:        " + app.Name)
	println("Slug:        " + app.Slug)
	println("Created by:  " + app.CreatedBy)
	println("Created at:  " + app.CreatedAt.Format(time.RFC1123))
	println("Description: " + app.Description)
	println("Status:      " + app.Status)
}
