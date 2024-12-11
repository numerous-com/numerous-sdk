package deletecmd

import (
	"context"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/mock"
)

var _ appDeleter = &mockAppDeleter{}

type mockAppDeleter struct {
	mock.Mock
}

// Delete implements AppService.
func (m *mockAppDeleter) Delete(ctx context.Context, input app.DeleteAppInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
