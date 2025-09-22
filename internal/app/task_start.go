package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type StartTaskInput struct {
	DeployID string
	TaskName string
}

type TaskStartResult struct {
	TaskInstanceID string
	TaskID         string
	Command        []string
}

type taskStartResponse struct {
	TaskStart taskStartResponseData `graphql:"taskStart(input: $input)"`
}

type taskStartResponseData struct {
	ID   string
	Task struct {
		ID      string
		Command []string
	}
}

const mutationTaskStartText = `
mutation CLITaskStart($input: TaskStartInput!) {
	taskStart(input: $input) {
		id
		task {
			id
			command
		}
	}
}
`

func (s *Service) StartTask(ctx context.Context, input StartTaskInput) (*TaskStartResult, error) {
	var resp taskStartResponse
	variables := map[string]any{
		"input": map[string]any{
			"deployID": graphql.ID(input.DeployID),
			"taskName": input.TaskName,
		},
	}

	err := s.client.Exec(ctx, mutationTaskStartText, &resp, variables, graphql.OperationName("CLITaskStart"))
	if err != nil {
		return nil, convertErrors(err)
	}

	result := taskStartFromResponse(resp.TaskStart)

	return &result, nil
}

func taskStartFromResponse(response taskStartResponseData) TaskStartResult {
	return TaskStartResult{
		TaskInstanceID: response.ID,
		TaskID:         response.Task.ID,
		Command:        response.Task.Command,
	}
}
