package revoke

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/token"
)

var _ TokenRevoker = &MockTokenRevoker{}

type MockTokenRevoker struct{ mock.Mock }

func (m *MockTokenRevoker) Revoke(ctx context.Context, id string) (token.RevokeTokenOutput, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(token.RevokeTokenOutput), args.Error(1)
}
