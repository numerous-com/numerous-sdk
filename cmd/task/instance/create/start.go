package create

import (
	"context"
	"errors"
	"os"
	"strings"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

var ErrConflictingInputFlags = errors.New("cannot specify both --input and --input-file")

type taskStartService interface {
	GetAppDeploymentID(ctx context.Context, organizationSlug, appSlug string) (string, error)
	StartTask(ctx context.Context, input app.StartTaskInput) (*app.TaskStartResult, error)
}

type TaskStartInput struct {
	AppDir           string
	OrganizationSlug string
	AppSlug          string
	TaskName         string
	Input            string
	InputFile        string
}

func startTask(ctx context.Context, service taskStartService, params TaskStartInput) error {
	ai, err := appident.GetAppIdentifier(params.AppDir, nil, params.OrganizationSlug, params.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, params.AppDir, ai)
		return err
	}

	if params.Input != "" && params.InputFile != "" {
		output.PrintError("Cannot specify both --input and --input-file", "")
		return ErrConflictingInputFlags
	}

	taskInput := params.Input
	if params.InputFile != "" {
		fileContent, err := os.ReadFile(params.InputFile)
		if err != nil {
			output.PrintErrorDetails("Error reading input file", err)
			return err
		}
		taskInput = string(fileContent)
	}

	deployID, err := service.GetAppDeploymentID(ctx, ai.OrganizationSlug, ai.AppSlug)
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	var inputPtr *string
	if taskInput != "" {
		inputPtr = &taskInput
	}

	result, err := service.StartTask(ctx, app.StartTaskInput{
		OrganizationSlug: ai.OrganizationSlug,
		DeployID:         deployID,
		TaskName:         params.TaskName,
		Input:            inputPtr,
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
