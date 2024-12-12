package status

import (
	"context"

	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/app"
)

var _ appReaderWorkloadLister = &mockAppReaderWorkloadLister{}

type mockAppReaderWorkloadLister struct{ mock.Mock }

// ListAppWorkloads implements appReaderWorkloadLister.
func (m *mockAppReaderWorkloadLister) ListAppWorkloads(ctx context.Context, input app.ListAppWorkloadsInput) ([]app.AppWorkload, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]app.AppWorkload), args.Error(1)
}

// ReadApp implements appReaderWorkloadLister.
func (m *mockAppReaderWorkloadLister) ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(app.ReadAppOutput), args.Error(1)
}
