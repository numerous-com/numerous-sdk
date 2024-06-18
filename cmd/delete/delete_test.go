package deleteapp

import (
	"testing"

	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestAppDelete(t *testing.T) {
	t.Run("returns nil and successfully sends AppDelete mutations", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		response, _ := test.DeleteSuccessQueryResult()
		app := app.App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
			PublicURL: "https://test.com/public/another-hash",
		}
		appQueryResponse := test.AppToQueryResult("tool", app)
		c, transportMock := test.CreateMockGqlClient(appQueryResponse, response)
		err := deleteApp(c, []string{})
		assert.NoError(t, err)
		transportMock.AssertExpectations(t)
	})
	t.Run("returns error if app does not exist", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		appNotFoundResponse := `"record not found"`
		response, _ := test.DeleteFailureQueryResult(appNotFoundResponse)
		c, transportMock := test.CreateMockGqlClient(response)

		err := deleteApp(c, []string{})

		assert.Error(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns error if app id document does not exists in the current directory", func(t *testing.T) {
		c, transportMock := test.CreateMockGqlClient()
		err := deleteApp(c, []string{})

		assert.Error(t, err)
		transportMock.AssertExpectations(t)
	})
}
