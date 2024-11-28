package share

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	appSlug := "app-slug"
	organizationSlug := "organization-slug"
	ai := appident.AppIdentifier{OrganizationSlug: organizationSlug, AppSlug: appSlug}
	sharedURL := "https://test-numerous.com/share/123"
	testErr := errors.New("test error")

	t.Run("it calls app lister with expected arguments", func(t *testing.T) {
		m := mockAppService{}
		m.On("ShareApp", ctx, ai).Once().Return(app.ShareAppOutput{SharedURL: &sharedURL}, nil)

		err := shareApp(ctx, &m, Input{AppDir: "", AppSlug: appSlug, OrgSlug: organizationSlug})

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("returns error if empty shared URL returned", func(t *testing.T) {
		m := mockAppService{}
		m.On("ShareApp", ctx, ai).Once().Return(app.ShareAppOutput{SharedURL: nil}, nil)

		err := shareApp(ctx, &m, Input{AppDir: "", AppSlug: appSlug, OrgSlug: organizationSlug})

		assert.ErrorIs(t, err, ErrEmptySharedURL)
	})

	t.Run("passes on error", func(t *testing.T) {
		for _, expectedError := range []error{
			app.ErrAccessDenied,
			app.ErrAppNotFound,
			testErr,
		} {
			t.Run(expectedError.Error(), func(t *testing.T) {
				m := mockAppService{}
				m.On("ShareApp", ctx, ai).Once().Return(app.ShareAppOutput{}, expectedError)

				err := shareApp(ctx, &m, Input{AppDir: "", AppSlug: appSlug, OrgSlug: organizationSlug})

				assert.ErrorIs(t, err, expectedError)
			})
		}
	})
}
