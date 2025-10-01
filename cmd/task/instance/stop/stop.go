package stop

import (
	"context"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type taskStopService interface {
	StopTask(ctx context.Context, taskInstanceID string) (*app.TaskStopResult, error)
}

type TaskStopInput struct {
	AppDir           string
	OrganizationSlug string
	AppSlug          string
	TaskInstanceID   string
}

func stopTask(ctx context.Context, service taskStopService, params TaskStopInput) error {
	ai, err := appident.GetAppIdentifier(params.AppDir, nil, params.OrganizationSlug, params.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, params.AppDir, ai)
		return err
	}

	result, err := service.StopTask(ctx, params.TaskInstanceID)
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	printTaskStopped(*result)

	return nil
}

func printTaskStopped(result app.TaskStopResult) {
	println("Task instance stopped.")
	println()
	println("Instance ID: " + result.TaskInstanceID)
}
