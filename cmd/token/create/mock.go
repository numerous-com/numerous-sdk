package create

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/token"
)

var _ TokenCreator = &MockTokenCreator{}

type MockTokenCreator struct{ mock.Mock }

func (m *MockTokenCreator) Create(ctx context.Context, input token.CreateTokenInput) (token.CreateTokenOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(token.CreateTokenOutput), args.Error(1)
}
