package download

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/app"
)

type mockAppService struct{ mock.Mock }

func (m *mockAppService) AppVersionDownloadURL(ctx context.Context, input app.AppVersionDownloadURLInput) (app.AppVersionDownloadURLOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.AppVersionDownloadURLOutput), args.Error(1)
}

func (m *mockAppService) CurrentAppVersion(ctx context.Context, input app.CurrentAppVersionInput) (app.CurrentAppVersionOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.CurrentAppVersionOutput), args.Error(1)
}
