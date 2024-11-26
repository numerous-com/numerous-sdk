package unpublish

import (
	"testing"

	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/test"

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

		c, transportMock := test.CreateMockGQLClient(appQueryResponse, appUnpublishResponse)
		err := unpublish(c)
		assert.NoError(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns error if app does not exist", func(t *testing.T) {
		test.ChdirToTmpDirWithAppIDDocument(t, dir.AppIDFileName, "id")
		appNotFoundResponse := `{"errors":[{"message":"record not found","path":["tool"]}],"data":null}`
		c, transportMock := test.CreateMockGQLClient(appNotFoundResponse)

		err := unpublish(c)

		assert.Error(t, err)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns error if app id document does not exists in the current directory", func(t *testing.T) {
		c, transportMock := test.CreateMockGQLClient()
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

		c, transportMock := test.CreateMockGQLClient(appQueryResponse)
		err := unpublish(c)

		assert.NoError(t, err)
		transportMock.AssertExpectations(t)
	})
}
