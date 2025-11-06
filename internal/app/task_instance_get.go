package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type GetTaskInstanceInput struct {
	TaskInstanceID string
}

type taskInstanceResponse struct {
	TaskInstance TaskInstance `graphql:"taskInstance(taskInstanceID: $taskInstanceID)"`
}

const queryTaskInstanceText = `
query CLIGetTaskInstance($taskInstanceID: ID!) {
	taskInstance(taskInstanceID: $taskInstanceID) {
		id
		createdAt
		task {
			id
			command
		}
		workload {
			status
			startedAt
			cpuUsage {
				current
			}
			memoryUsageMB {
				current
			}
			exitCode
			input
		}
	}
}
`

func (s *Service) GetTaskInstance(ctx context.Context, input GetTaskInstanceInput) (*TaskInstance, error) {
	var resp taskInstanceResponse
	variables := map[string]any{
		"taskInstanceID": graphql.ID(input.TaskInstanceID),
	}

	err := s.client.Exec(ctx, queryTaskInstanceText, &resp, variables, graphql.OperationName("CLIGetTaskInstance"))
	if err != nil {
		return nil, convertErrors(err)
	}

	return &resp.TaskInstance, nil
}
