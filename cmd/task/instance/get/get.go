package get

import (
	"context"
	"fmt"
	"strings"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/output"
)

type taskGetService interface {
	GetTaskInstance(ctx context.Context, input app.GetTaskInstanceInput) (*app.TaskInstance, error)
}

type TaskGetInput struct {
	TaskInstanceID string
}

func getInstance(ctx context.Context, service taskGetService, params TaskGetInput) error {
	taskInstance, err := service.GetTaskInstance(ctx, app.GetTaskInstanceInput{
		TaskInstanceID: params.TaskInstanceID,
	})
	if err != nil {
		output.PrintErrorDetails("Error retrieving task instance", err)
		return err
	}

	printTaskInstanceDetails(*taskInstance)

	return nil
}

func printTaskInstanceDetails(taskInstance app.TaskInstance) {
	commandStr := strings.Join(taskInstance.Task.Command, " ")

	println("Task Instance Details:")
	println()
	println("Instance: " + taskInstance.ID)
	println("Task:     " + taskInstance.Task.ID)
	println("Status:   " + taskInstance.Workload.Status)
	println("Created:  " + taskInstance.CreatedAt.Format(time.RFC3339))
	println("Command:  " + commandStr)

	if taskInstance.Input != nil {
		decodedInput := app.DecodeTaskDataForDisplay(taskInstance.Input)
		println("Input:    " + decodedInput)
	}

	if taskInstance.Output != nil {
		decodedOutput := app.DecodeTaskDataForDisplay(taskInstance.Output)
		println("Output:   " + decodedOutput)
	}

	if taskInstance.Progress.Value != nil {
		progressStr := fmt.Sprintf("Progress: %.1f", *taskInstance.Progress.Value)
		if taskInstance.Progress.Message != nil && *taskInstance.Progress.Message != "" {
			progressStr += fmt.Sprintf(" (%s)", *taskInstance.Progress.Message)
		}
		println(progressStr)
	}

	if taskInstance.Workload.ExitCode != nil {
		println(fmt.Sprintf("ExitCode: %d", *taskInstance.Workload.ExitCode))
	}

	if taskInstance.Workload.CPUUsage != nil && taskInstance.Workload.MemoryUsageMB != nil {
		println(fmt.Sprintf("CPU:      %.1f", taskInstance.Workload.CPUUsage.Current))
		println(fmt.Sprintf("Memory:   %.0fMB", taskInstance.Workload.MemoryUsageMB.Current))
	}
}
