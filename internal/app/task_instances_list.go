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
}

type WorkloadResourceUsage struct {
	Current float64
}

type ListTaskInstancesInput struct {
	DeployID string
}

type taskInstancesResponse struct {
	TaskInstances []taskInstanceResponseData `graphql:"taskInstances(deployID: $deployID)"`
}

type taskInstanceResponseData struct {
	ID        string
	CreatedAt time.Time
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
	}
}

const queryTaskInstancesText = `
query CLIListTaskInstances($deployID: ID!) {
	taskInstances(deployID: $deployID) {
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
		}
	}
}
`

func (s *Service) ListTaskInstances(ctx context.Context, input ListTaskInstancesInput) ([]TaskInstance, error) {
	var resp taskInstancesResponse
	variables := map[string]any{"deployID": input.DeployID}

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

func taskInstanceFromResponse(responseTaskInstance taskInstanceResponseData) TaskInstance {
	ti := TaskInstance{
		ID:        responseTaskInstance.ID,
		CreatedAt: responseTaskInstance.CreatedAt,
		Task: Task{
			ID:      responseTaskInstance.Task.ID,
			Command: responseTaskInstance.Task.Command,
		},
		Workload: TaskInstanceWorkload{
			Status:    responseTaskInstance.Workload.Status,
			StartedAt: responseTaskInstance.Workload.StartedAt,
		},
	}

	if responseTaskInstance.Workload.CPUUsage != nil {
		ti.Workload.CPUUsage = &WorkloadResourceUsage{
			Current: responseTaskInstance.Workload.CPUUsage.Current,
		}
	}

	if responseTaskInstance.Workload.MemoryUsageMB != nil {
		ti.Workload.MemoryUsageMB = &WorkloadResourceUsage{
			Current: responseTaskInstance.Workload.MemoryUsageMB.Current,
		}
	}

	return ti
}
