package unshare

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

func TestUnshare(t *testing.T) {
	ctx := context.TODO()
	appSlug := "app-slug"
	organizationSlug := "organization-slug"
	ai := appident.AppIdentifier{OrganizationSlug: organizationSlug, AppSlug: appSlug}
	testErr := errors.New("test error")

	t.Run("it calls app service with expected arguments", func(t *testing.T) {
		m := mockAppService{}
		m.On("UnshareApp", ctx, ai).Once().Return(nil)

		err := unshareApp(ctx, &m, Input{AppDir: "", AppSlug: appSlug, OrgSlug: organizationSlug})

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("passes on error", func(t *testing.T) {
		for _, expectedError := range []error{
			app.ErrAccessDenied,
			app.ErrAppNotFound,
			testErr,
		} {
			t.Run(expectedError.Error(), func(t *testing.T) {
				m := mockAppService{}
				m.On("UnshareApp", ctx, ai).Once().Return(expectedError)

				err := unshareApp(ctx, &m, Input{AppDir: "", AppSlug: appSlug, OrgSlug: organizationSlug})

				assert.ErrorIs(t, err, expectedError)
			})
		}
	})
}
