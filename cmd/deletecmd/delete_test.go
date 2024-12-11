package deletecmd

import (
	"context"
	"errors"
	"testing"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	const appSlug = "app-slug"
	const slug = "organization-slug"
	testError := errors.New("test error")

	t.Run("given app with manifest then it deletes expected app", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		service := &mockAppDeleter{}

		expectedInput := app.DeleteAppInput{OrganizationSlug: "organization-slug-in-manifest", AppSlug: "app-slug-in-manifest"}
		service.On("Delete", mock.Anything, expectedInput).Return(nil)

		err := deleteApp(context.TODO(), service, appDir, "", "")
		assert.NoError(t, err)
	})

	t.Run("given slug and app slug arguments then it deletes expected app", func(t *testing.T) {
		appDir := t.TempDir()
		service := &mockAppDeleter{}

		expectedInput := app.DeleteAppInput{OrganizationSlug: slug, AppSlug: appSlug}
		service.On("Delete", mock.Anything, expectedInput).Return(nil)

		err := deleteApp(context.TODO(), service, appDir, slug, appSlug)
		assert.NoError(t, err)
	})

	t.Run("given error is returned then it passes the error on", func(t *testing.T) {
		appDir := t.TempDir()
		service := &mockAppDeleter{}

		expectedInput := app.DeleteAppInput{OrganizationSlug: slug, AppSlug: appSlug}
		service.On("Delete", mock.Anything, expectedInput).Return(testError)

		err := deleteApp(context.TODO(), service, appDir, slug, appSlug)

		assert.ErrorIs(t, err, testError)
	})
}
