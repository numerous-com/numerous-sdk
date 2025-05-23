package app

import (
	"testing"
	"time"

	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	m := &manifest.Manifest{
		App: manifest.App{
			Name:       "name",
			CoverImage: "cover.png",
		},
		Python: &manifest.Python{
			Library:          manifest.LibraryMarimo,
			Version:          "3.11",
			AppFile:          "app.py",
			RequirementsFile: "requirements.txt",
		},
	}
	t.Run("can return app on AppCreate mutation", func(t *testing.T) {
		expectedApp := App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
			PublicURL: "https://test.com/public/another-hash",
			Name:      "test name",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		response := test.AppToQueryResult("toolCreate", expectedApp)
		c := test.CreateTestGqlClient(t, response)

		actualApp, err := Create(m, c)

		assert.NoError(t, err)
		assert.Equal(t, expectedApp, actualApp)
	})

	t.Run("can return error on AppCreate mutation", func(t *testing.T) {
		appNotFoundResponse := `{"errors":[{"message":"Something went wrong","path":["toolCreate"]}],"data":null}`
		c := test.CreateTestGqlClient(t, appNotFoundResponse)

		actualApp, err := Create(m, c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Something went wrong")
		assert.Equal(t, App{}, actualApp)
	})
}
