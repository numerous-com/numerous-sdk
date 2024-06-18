package logs

import (
	"numerous/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

var _ AppService = &AppServiceMock{}

type AppServiceMock struct {
	mock.Mock
}

// AppDeployLogs implements AppService.
func (m *AppServiceMock) AppDeployLogs(slug string, appName string) (chan app.AppDeployLogEntry, error) {
	args := m.Called(slug, appName)
	return args.Get(0).(chan app.AppDeployLogEntry), args.Error(1)
}
