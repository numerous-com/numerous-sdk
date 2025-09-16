package logs

import (
	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

type TaskLogsServiceMock struct {
	mock.Mock
}

// TaskInstanceLogs implements taskLogsService.
func (m *TaskLogsServiceMock) TaskInstanceLogs(input app.TaskInstanceLogsInput) (chan app.WorkloadLogEntry, error) {
	args := m.Called(input)
	return args.Get(0).(chan app.WorkloadLogEntry), args.Error(1)
}
