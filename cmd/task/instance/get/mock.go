package get

import (
	"context"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

type TaskGetServiceMock struct {
	mock.Mock
}

func (m *TaskGetServiceMock) GetTaskInstance(ctx context.Context, input app.GetTaskInstanceInput) (*app.TaskInstance, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*app.TaskInstance), args.Error(1)
}
