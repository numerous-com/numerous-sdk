package unshare

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/appident"
)

type mockAppService struct{ mock.Mock }

var _ AppService = &mockAppService{}

func (m *mockAppService) UnshareApp(ctx context.Context, ai appident.AppIdentifier) error {
	args := m.Called(ctx, ai)
	return args.Error(0)
}
