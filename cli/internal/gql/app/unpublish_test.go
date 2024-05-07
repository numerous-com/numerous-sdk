package app

import (
	"testing"
	"time"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestUnpublish(t *testing.T) {
	t.Run("can return app on AppUnpublish mutation", func(t *testing.T) {
		expectedApp := App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
			PublicURL: "",
			Name:      "test name",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		response := test.AppToQueryResult("toolUnpublish", expectedApp)
		c := test.CreateTestGqlClient(t, response)

		actualApp, err := Unpublish(expectedApp.ID, c)

		assert.NoError(t, err)
		assert.Equal(t, expectedApp, actualApp)
	})

	t.Run("can return error on AppUnpublish mutation", func(t *testing.T) {
		appNotFoundResponse := `{"errors":[{"message":"record not found","path":["toolUnpublish"]}],"data":null}`
		c := test.CreateTestGqlClient(t, appNotFoundResponse)

		actualApp, err := Unpublish("id", c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "record not found")
		assert.Equal(t, App{}, actualApp)
	})
}
