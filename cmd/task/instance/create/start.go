package create

import (
	"context"
	"strings"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type taskStartService interface {
	GetAppDeploymentID(ctx context.Context, organizationSlug, appSlug string) (string, error)
	StartTask(ctx context.Context, input app.StartTaskInput) (*app.TaskStartResult, error)
}

type TaskStartInput struct {
	AppDir           string
	OrganizationSlug string
	AppSlug          string
	TaskName         string
}

func startTask(ctx context.Context, service taskStartService, params TaskStartInput) error {
	ai, err := appident.GetAppIdentifier(params.AppDir, nil, params.OrganizationSlug, params.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, params.AppDir, ai)
		return err
	}

	deployID, err := service.GetAppDeploymentID(ctx, ai.OrganizationSlug, ai.AppSlug)
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	result, err := service.StartTask(ctx, app.StartTaskInput{
		DeployID: deployID,
		TaskName: params.TaskName,
	})
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	printTaskStarted(*result)

	return nil
}

func printTaskStarted(result app.TaskStartResult) {
	commandStr := strings.Join(result.Command, " ")

	println("Task instance created.")
	println()
	println("Instance: " + result.TaskInstanceID)
	println("Task:     " + result.TaskID)
	println("Command:  " + commandStr)
}
