package logs

import (
	"numerous.com/cli/cmd/app/appident"
	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

var _ AppService = &AppServiceMock{}

type AppServiceMock struct {
	mock.Mock
}

// AppDeployLogs implements AppService.
func (m *AppServiceMock) AppDeployLogs(ai appident.AppIdentifier) (chan app.AppDeployLogEntry, error) {
	args := m.Called(ai)
	return args.Get(0).(chan app.AppDeployLogEntry), args.Error(1)
}
