package version

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/version"
)

var _ VersionChecker = &MockVersionChecker{}

type MockVersionChecker struct{ mock.Mock }

func (m *MockVersionChecker) Check(ctx context.Context) (version.CheckVersionOutput, error) {
	args := m.Called(ctx)
	return args.Get(0).(version.CheckVersionOutput), args.Error(1)
}
