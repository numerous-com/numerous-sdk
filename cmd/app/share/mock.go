package share

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type mockAppService struct{ mock.Mock }

var _ AppService = &mockAppService{}

func (m *mockAppService) ShareApp(ctx context.Context, ai appident.AppIdentifier) (app.ShareAppOutput, error) {
	args := m.Called(ctx, ai)
	return args.Get(0).(app.ShareAppOutput), args.Error(1)
}
