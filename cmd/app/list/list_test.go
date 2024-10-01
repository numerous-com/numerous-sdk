package list

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/app"
)

var _ AppLister = &mockAppLister{}

type mockAppLister struct{ mock.Mock }

func (m *mockAppLister) List(ctx context.Context, organizationSlug string) ([]app.ListApp, error) {
	args := m.Called(ctx, organizationSlug)
	return args.Get(0).([]app.ListApp), args.Error(1)
}

var errTest = errors.New("test error")

func TestList(t *testing.T) {
	ctx := context.TODO()
	organizationSlug := "organization-slug"

	la1 := app.ListApp{
		Name:        "App 1",
		Slug:        "app-1",
		Description: "App 1 description",
		Status:      "RUNNING",
		CreatedBy:   "App 1 User",
		CreatedAt:   time.Date(2024, time.March, 29, 12, 12, 12, 0, time.UTC),
	}

	la2 := app.ListApp{
		Name:        "App 2",
		Slug:        "app-2",
		Description: "App 2 description",
		Status:      "RUNNING",
		CreatedBy:   "App 2 User",
		CreatedAt:   time.Date(2024, time.March, 29, 22, 22, 22, 0, time.UTC),
	}

	t.Run("it calls app lister with expected arguments", func(t *testing.T) {
		m := mockAppLister{}
		m.On("List", ctx, organizationSlug).Once().Return([]app.ListApp{la1, la2}, nil)

		err := list(ctx, &m, AppListInput{OrganizationSlug: organizationSlug})

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("it returns error from app lister", func(t *testing.T) {
		m := mockAppLister{}
		m.On("List", ctx, organizationSlug).Once().Return(([]app.ListApp)(nil), errTest)

		err := list(ctx, &m, AppListInput{OrganizationSlug: organizationSlug})

		assert.ErrorIs(t, err, errTest)
		m.AssertExpectations(t)
	})
}
