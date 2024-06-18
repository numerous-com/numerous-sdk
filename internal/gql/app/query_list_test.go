package app

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestQueryList(t *testing.T) {
	t.Run("can return apps on list apps query", func(t *testing.T) {
		appsAsStrings := []string{}
		expectedApps := []App{}
		for i := 0; i < 3; i++ {
			id := strconv.Itoa(i)
			app := App{
				ID:        id,
				SharedURL: "https://test.com/shared/some-hash-" + id,
				PublicURL: "",
				Name:      "Name " + id,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			expectedApps = append(expectedApps, app)
			appsAsStrings = append(appsAsStrings, test.AppToResponse(app))
		}

		response := `{
			"data": {
				"tools": [` + strings.Join(appsAsStrings, ", ") + `]
			}
		}`
		c := test.CreateTestGqlClient(t, response)

		actualApps, err := QueryList(c)

		assert.NoError(t, err)
		assert.Equal(t, expectedApps, actualApps)
	})

	t.Run("can return permission denied error", func(t *testing.T) {
		appNotFoundResponse := `{"errors":[{"message":"permission denied","path":["tools"]}],"data":null}`
		c := test.CreateTestGqlClient(t, appNotFoundResponse)

		actualApps, err := QueryList(c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "permission denied")
		assert.Equal(t, []App{}, actualApps)
	})
}
