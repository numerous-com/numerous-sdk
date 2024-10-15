package deploy

import (
	"context"
	"io"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"

	"github.com/stretchr/testify/mock"
)

var _ AppService = &mockAppService{}

type mockAppService struct {
	mock.Mock
}

// ReadApp implements AppService.
func (m *mockAppService) ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.ReadAppOutput), args.Error(1)
}

// DeployEvents implements AppService.
func (m *mockAppService) DeployEvents(ctx context.Context, input app.DeployEventsInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// DeployApp implements AppService.
func (m *mockAppService) DeployApp(ctx context.Context, input app.DeployAppInput) (app.DeployAppOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.DeployAppOutput), args.Error(1)
}

// AppVersionUploadURL implements AppService.
func (m *mockAppService) AppVersionUploadURL(ctx context.Context, input app.AppVersionUploadURLInput) (app.AppVersionUploadURLOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.AppVersionUploadURLOutput), args.Error(1)
}

// Create implements AppService.
func (m *mockAppService) Create(ctx context.Context, input app.CreateAppInput) (app.CreateAppOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.CreateAppOutput), args.Error(1)
}

// CreateVersion implements AppService.
func (m *mockAppService) CreateVersion(ctx context.Context, input app.CreateAppVersionInput) (app.CreateAppVersionOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.CreateAppVersionOutput), args.Error(1)
}

// UploadAppSource implements AppService.
func (m *mockAppService) UploadAppSource(uploadURL string, archive io.Reader) error {
	args := m.Called(uploadURL, archive)
	return args.Error(0)
}

// AppDeployLogs implements AppService.
func (m *mockAppService) AppDeployLogs(ai appident.AppIdentifier) (chan app.AppDeployLogEntry, error) {
	args := m.Called(ai)
	return args.Get(0).(chan app.AppDeployLogEntry), args.Error(1)
}
