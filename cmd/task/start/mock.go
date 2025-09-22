package start

import (
	"context"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

type TaskStartServiceMock struct {
	mock.Mock
}

// GetAppDeploymentID implements taskStartService.
func (m *TaskStartServiceMock) GetAppDeploymentID(ctx context.Context, organizationSlug, appSlug string) (string, error) {
	args := m.Called(ctx, organizationSlug, appSlug)
	return args.String(0), args.Error(1)
}

// StartTask implements taskStartService.
func (m *TaskStartServiceMock) StartTask(ctx context.Context, input app.StartTaskInput) (*app.TaskStartResult, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*app.TaskStartResult), args.Error(1)
}
