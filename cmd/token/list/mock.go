package list

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/token"
)

var _ TokenLister = &MockTokenLister{}

type MockTokenLister struct{ mock.Mock }

func (m *MockTokenLister) List(ctx context.Context) (token.ListTokenOutput, error) {
	args := m.Called(ctx)
	return args.Get(0).(token.ListTokenOutput), args.Error(1)
}
