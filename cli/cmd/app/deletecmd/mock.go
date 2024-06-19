package deletecmd

import (
	"context"

	"numerous/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

var _ AppService = &MockAppService{}

type MockAppService struct {
	mock.Mock
}

// Delete implements AppService.
func (m *MockAppService) Delete(ctx context.Context, input app.DeleteAppInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
