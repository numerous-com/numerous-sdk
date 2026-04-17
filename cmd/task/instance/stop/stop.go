package stop

import (
	"context"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/output"
)

type taskStopService interface {
	StopTask(ctx context.Context, taskInstanceID string) (*app.TaskStopResult, error)
}

type TaskStopInput struct {
	TaskInstanceID string
}

func stopTask(ctx context.Context, service taskStopService, params TaskStopInput) error {
	result, err := service.StopTask(ctx, params.TaskInstanceID)
	if err != nil {
		output.PrintErrorDetails("Error stopping task instance", err)
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
