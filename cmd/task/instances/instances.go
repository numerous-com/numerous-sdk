package instances

import (
	"context"
	"fmt"
	"strings"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type taskInstancesService interface {
	GetAppDeploymentID(ctx context.Context, organizationSlug, appSlug string) (string, error)
	ListTaskInstances(ctx context.Context, input app.ListTaskInstancesInput) ([]app.TaskInstance, error)
}

type TaskInstancesInput struct {
	AppDir           string
	OrganizationSlug string
	AppSlug          string
	TaskID           string
}

func listInstances(ctx context.Context, service taskInstancesService, params TaskInstancesInput) error {
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

	taskInstances, err := service.ListTaskInstances(ctx, app.ListTaskInstancesInput{
		DeployID: deployID,
		TaskID:   params.TaskID,
	})
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	if len(taskInstances) == 0 {
		println(fmt.Sprintf("No instances found for task '%s'.", params.TaskID))
		return nil
	}

	println(fmt.Sprintf("Task Instances (%s):", params.TaskID))

	for i, instance := range taskInstances {
		if i > 0 {
			println()
		}
		printTaskInstance(instance)
	}

	return nil
}

func printTaskInstance(taskInstance app.TaskInstance) {
	commandStr := strings.Join(taskInstance.Task.Command, " ")

	println("Instance: " + taskInstance.ID)
	println("Task:     " + taskInstance.Task.ID)
	println("Status:   " + taskInstance.Workload.Status)
	println("Created:  " + taskInstance.CreatedAt.Format(time.RFC3339))
	println("Command:  " + commandStr)

	if taskInstance.Workload.ExitCode != nil {
		println(fmt.Sprintf("ExitCode: %d", *taskInstance.Workload.ExitCode))
	}

	if taskInstance.Workload.CPUUsage != nil && taskInstance.Workload.MemoryUsageMB != nil {
		println(fmt.Sprintf("CPU:      %.1f", taskInstance.Workload.CPUUsage.Current))
		println(fmt.Sprintf("Memory:   %.0fMB", taskInstance.Workload.MemoryUsageMB.Current))
	}
}
