package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type TaskStopResult struct {
	TaskInstanceID string
}

type taskStopResponse struct {
	TaskStop struct {
		TaskInstanceID string `graphql:"taskInstanceID"`
	} `graphql:"taskStop(taskInstanceID: $taskInstanceID)"`
}

const mutationTaskStopText = `
mutation CLITaskStop($taskInstanceID: ID!) {
	taskStop(taskInstanceID: $taskInstanceID) {
		taskInstanceID
	}
}
`

func (s *Service) StopTask(ctx context.Context, taskInstanceID string) (*TaskStopResult, error) {
	var resp taskStopResponse
	variables := map[string]any{
		"taskInstanceID": graphql.ID(taskInstanceID),
	}

	err := s.client.Exec(ctx, mutationTaskStopText, &resp, variables, graphql.OperationName("CLITaskStop"))
	if err != nil {
		return nil, convertErrors(err)
	}

	result := TaskStopResult{
		TaskInstanceID: resp.TaskStop.TaskInstanceID,
	}

	return &result, nil
}
