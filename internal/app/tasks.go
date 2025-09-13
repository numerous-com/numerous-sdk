package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type GetTasksInput struct {
	OrganizationSlug string
	AppSlug          string
}

type tasksResponse struct {
	App struct {
		ID                string
		DefaultDeployment *struct {
			Current *struct {
				AppVersion struct {
					Tasks []struct {
						ID      string
						Command []string
					}
				}
			}
		}
	}
}

const queryTasksText = `
query CLIGetTasks($orgSlug: String!, $appSlug: String!) {
	app(organizationSlug: $orgSlug, appSlug: $appSlug) {
		id
		defaultDeployment {
			current {
				appVersion {
					tasks {
						id
						command
					}
				}
			}
		}
	}
}
`

func (s *Service) GetTasks(ctx context.Context, input GetTasksInput) ([]Task, error) {
	var resp tasksResponse
	variables := map[string]any{"orgSlug": input.OrganizationSlug, "appSlug": input.AppSlug}

	err := s.client.Exec(ctx, queryTasksText, &resp, variables, graphql.OperationName("CLIGetTasks"))
	if err != nil {
		return nil, convertErrors(err)
	}

	if resp.App.DefaultDeployment == nil || resp.App.DefaultDeployment.Current == nil {
		return nil, ErrDeploymentNotFound
	}

	var tasks []Task
	for _, task := range resp.App.DefaultDeployment.Current.AppVersion.Tasks {
		tasks = append(tasks, Task{
			ID:      task.ID,
			Command: task.Command,
		})
	}

	return tasks, nil
}
