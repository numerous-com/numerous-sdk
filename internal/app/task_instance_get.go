package app

import (
	"context"
	"errors"

	"github.com/hasura/go-graphql-client"
)

type GetTaskInstanceInput struct {
	TaskInstanceID string
}

var ErrTaskInstanceNotFound = errors.New("task instance not found")

type taskInstanceResponse struct {
	TaskInstance *TaskInstance `graphql:"taskInstance(taskInstanceID: $taskInstanceID)"`
}

const queryTaskInstanceText = `
query CLIGetTaskInstance($taskInstanceID: ID!) {
	taskInstance(taskInstanceID: $taskInstanceID) {
		id
		createdAt
		input
		output
		progress {
			value
			message
		}
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

	if resp.TaskInstance == nil {
		return nil, ErrTaskInstanceNotFound
	}

	return resp.TaskInstance, nil
}
