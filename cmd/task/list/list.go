package list

import (
	"context"
	"strings"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type tasksService interface {
	GetTasks(ctx context.Context, input app.GetTasksInput) ([]app.Task, error)
}

type TaskListInput struct {
	AppDir           string
	OrganizationSlug string
	AppSlug          string
}

func list(ctx context.Context, service tasksService, params TaskListInput) error {
	ai, err := appident.GetAppIdentifier(params.AppDir, nil, params.OrganizationSlug, params.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, params.AppDir, ai)
		return err
	}

	tasks, err := service.GetTasks(ctx, app.GetTasksInput{
		OrganizationSlug: ai.OrganizationSlug,
		AppSlug:          ai.AppSlug,
	})
	if err != nil {
		app.PrintTaskError(err, ai)
		return err
	}

	if len(tasks) == 0 {
		println("No tasks defined in this app version.")
		return nil
	}

	println("Tasks:")
	for i, task := range tasks {
		if i > 0 {
			println()
		}
		printTask(task)
	}

	return nil
}

func printTask(task app.Task) {
	commandStr := strings.Join(task.Command, " ")

	println("ID:      " + task.ID)
	println("Command: " + commandStr)
}
