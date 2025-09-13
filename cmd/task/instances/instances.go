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
	})
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	filteredInstances := filterInstancesByTaskID(taskInstances, params.TaskID)

	if len(filteredInstances) == 0 {
		println(fmt.Sprintf("No instances found for task '%s'.", params.TaskID))
		return nil
	}

	println(fmt.Sprintf("Task Instances (%s):", params.TaskID))

	for i, instance := range filteredInstances {
		if i > 0 {
			println()
		}
		printTaskInstance(instance)
	}

	return nil
}

func filterInstancesByTaskID(instances []app.TaskInstance, taskID string) []app.TaskInstance {
	var filtered []app.TaskInstance
	for _, instance := range instances {
		if strings.EqualFold(instance.Task.ID, taskID) {
			filtered = append(filtered, instance)
		}
	}

	return filtered
}

func printTaskInstance(taskInstance app.TaskInstance) {
	commandStr := strings.Join(taskInstance.Task.Command, " ")

	println("Instance: " + taskInstance.ID)
	println("Task:     " + taskInstance.Task.ID)
	println("Status:   " + taskInstance.Workload.Status)
	println("Duration: " + getDurationStr(taskInstance.Workload.StartedAt))
	println("Command:  " + commandStr)

	if taskInstance.Workload.CPUUsage != nil && taskInstance.Workload.MemoryUsageMB != nil {
		println(fmt.Sprintf("CPU:      %.1f", taskInstance.Workload.CPUUsage.Current))
		println(fmt.Sprintf("Memory:   %.0fMB", taskInstance.Workload.MemoryUsageMB.Current))
	}
}

func getDurationStr(startedAt time.Time) string {
	duration := time.Since(startedAt)

	switch {
	case duration.Hours() >= 1:
		return fmt.Sprintf("%.0fh", duration.Hours())
	case duration.Minutes() >= 1:
		return fmt.Sprintf("%.0fm", duration.Minutes())
	default:
		return fmt.Sprintf("%.0fs", duration.Seconds())
	}
}
