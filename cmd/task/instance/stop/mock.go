package stop

import (
	"context"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

type TaskStopServiceMock struct {
	mock.Mock
}

// StopTask implements taskStopService.
func (m *TaskStopServiceMock) StopTask(ctx context.Context, taskInstanceID string) (*app.TaskStopResult, error) {
	args := m.Called(ctx, taskInstanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*app.TaskStopResult), args.Error(1)
}
