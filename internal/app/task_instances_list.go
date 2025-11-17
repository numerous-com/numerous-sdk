package app

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
)

type TaskInstance struct {
	ID        string
	Task      Task
	CreatedAt time.Time
	Input     *string
	Output    *string
	Workload  TaskInstanceWorkload
}

type Task struct {
	ID      string
	Command []string
}

type TaskInstanceWorkload struct {
	Status        string
	StartedAt     time.Time
	CPUUsage      *WorkloadResourceUsage
	MemoryUsageMB *WorkloadResourceUsage
	ExitCode      *int
}

type WorkloadResourceUsage struct {
	Current float64
}

type ListTaskInstancesInput struct {
	OrganizationSlug string
	DeployID         string
	TaskID           string
}

type taskInstancesResponse struct {
	TaskInstances []taskInstanceResponseData `graphql:"taskInstances(organizationSlug: $organizationSlug, deployID: $deployID, taskID: $taskID)"`
}

type taskInstanceResponseData struct {
	ID        string
	CreatedAt time.Time
	Input     *string
	Output    *string
	Task      struct {
		ID      string
		Command []string
	}
	Workload struct {
		Status    string
		StartedAt time.Time
		CPUUsage  *struct {
			Current float64
		}
		MemoryUsageMB *struct {
			Current float64
		}
		ExitCode *int
	}
}

const queryTaskInstancesText = `
query CLIListTaskInstances($organizationSlug: String!, $deployID: ID!, $taskID: ID!) {
	taskInstances(organizationSlug: $organizationSlug, deployID: $deployID, taskID: $taskID) {
		id
		createdAt
		input
		output
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

func (s *Service) ListTaskInstances(ctx context.Context, input ListTaskInstancesInput) ([]TaskInstance, error) {
	var resp taskInstancesResponse
	variables := map[string]any{
		"organizationSlug": input.OrganizationSlug,
		"deployID":         graphql.ID(input.DeployID),
		"taskID":           graphql.ID(input.TaskID),
	}

	err := s.client.Exec(ctx, queryTaskInstancesText, &resp, variables, graphql.OperationName("CLIListTaskInstances"))
	if err != nil {
		return nil, convertErrors(err)
	}

	taskInstances := []TaskInstance{}
	for _, ti := range resp.TaskInstances {
		taskInstances = append(taskInstances, taskInstanceFromResponse(ti))
	}

	return taskInstances, nil
}

func taskInstanceFromResponse(response taskInstanceResponseData) TaskInstance {
	ti := TaskInstance{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		Input:     response.Input,
		Output:    response.Output,
		Task: Task{
			ID:      response.Task.ID,
			Command: response.Task.Command,
		},
		Workload: TaskInstanceWorkload{
			Status:    response.Workload.Status,
			StartedAt: response.Workload.StartedAt,
			ExitCode:  response.Workload.ExitCode,
		},
	}

	if response.Workload.CPUUsage != nil {
		ti.Workload.CPUUsage = &WorkloadResourceUsage{
			Current: response.Workload.CPUUsage.Current,
		}
	}

	if response.Workload.MemoryUsageMB != nil {
		ti.Workload.MemoryUsageMB = &WorkloadResourceUsage{
			Current: response.Workload.MemoryUsageMB.Current,
		}
	}

	return ti
}
