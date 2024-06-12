package unpublish

import (
	"testing"

	"numerous/cli/internal/dir"
	"numerous/cli/internal/gql/app"
	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestAppPublish(t *testing.T) {
	t.Run("returns nil and successfully sends AppUnpublish mutations", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		app := app.App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
			PublicURL: "https://test.com/public/another-hash",
		}
		appQueryResponse := test.AppToQueryResult("tool", app)
		test.AppToQueryResult("ad", app)
		app.PublicURL = ""
		appUnpublishResponse := test.AppToQueryResult("toolUnpublish", app)

		c, transportMock := test.CreateMockGqlClient(appQueryResponse, appUnpublishResponse)
		err := unpublish(c)
		assert.NoError(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns error if app does not exist", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		appNotFoundResponse := `{"errors":[{"message":"record not found","path":["tool"]}],"data":null}`
		c, transportMock := test.CreateMockGqlClient(appNotFoundResponse)

		err := unpublish(c)

		assert.Error(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns error if app id document does not exists in the current directory", func(t *testing.T) {
		c, transportMock := test.CreateMockGqlClient()
		err := unpublish(c)

		assert.Error(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("return nil and does not send AppUnpublish mutation, if app is not published", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		app := app.App{
			ID:        "id",
			SharedURL: "https://test.com/shared/some-hash",
		}
		appQueryResponse := test.AppToQueryResult("tool", app)

		c, transportMock := test.CreateMockGqlClient(appQueryResponse)
		err := unpublish(c)

		assert.NoError(t, err)
		transportMock.AssertExpectations(t)
	})
}
