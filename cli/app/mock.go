package app

import (
	"context"
	"io"

	"numerous/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

var _ AppService = &mockAppService{}

type mockAppService struct {
	mock.Mock
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
