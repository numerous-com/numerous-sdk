package app

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/hasura/go-graphql-client"
)

const (
	MaxTaskInputSize = 4096
)

var ErrTaskInputTooLarge = errors.New("task input too large: maximum size is 4KB (base64-encoded)")

type StartTaskInput struct {
	DeployID string
	TaskName string
	Input    *string
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

	taskInput := map[string]any{
		"deployID": graphql.ID(input.DeployID),
		"taskName": input.TaskName,
	}

	if input.Input != nil {
		encodedInput, err := encodeTaskInput(*input.Input)
		if err != nil {
			return nil, err
		}
		taskInput["input"] = encodedInput
	}

	variables := map[string]any{
		"input": taskInput,
	}

	err := s.client.Exec(ctx, mutationTaskStartText, &resp, variables, graphql.OperationName("CLITaskStart"))
	if err != nil {
		return nil, convertErrors(err)
	}

	result := taskStartFromResponse(resp.TaskStart)

	return &result, nil
}

func encodeTaskInput(rawInput string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(rawInput))
	if len(encoded) > MaxTaskInputSize {
		return "", ErrTaskInputTooLarge
	}

	return encoded, nil
}

func taskStartFromResponse(response taskStartResponseData) TaskStartResult {
	return TaskStartResult{
		TaskInstanceID: response.ID,
		TaskID:         response.Task.ID,
		Command:        response.Task.Command,
	}
}
