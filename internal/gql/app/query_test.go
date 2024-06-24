package app

import (
	"testing"
	"time"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	t.Run("can return app on app query", func(t *testing.T) {
		expectedApp := App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
			PublicURL: "",
			Name:      "test name",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		response := test.AppToQueryResult("tool", expectedApp)
		c := test.CreateTestGqlClient(t, response)

		actualApp, err := Query("id", c)

		assert.NoError(t, err)
		assert.Equal(t, expectedApp, actualApp)
	})

	t.Run("can return error on app query", func(t *testing.T) {
		appNotFoundResponse := `{"errors":[{"message":"record not found","path":["tool"]}],"data":null}`
		c := test.CreateTestGqlClient(t, appNotFoundResponse)

		actualApp, err := Query("non-existing-id", c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "record not found")
		assert.Equal(t, App{}, actualApp)
	})
}
