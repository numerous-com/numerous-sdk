package build

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"numerous/cli/test"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	tmpFilePath := filepath.Join(t.TempDir(), "file.zip")
	tmpFile, err := os.Create(tmpFilePath)
	require.NoError(t, err)
	defer tmpFile.Close()

	t.Run("can return app on AppCreate mutation", func(t *testing.T) {
		expectedBuild := BuildConfiguration{
			BuildID: "buildID",
		}
		response := `{
			"data": {
				"buildPush": {
					"buildId": "buildID"
				}
			}
		}`
		c := createTestGqlClient(response)
		appID := "app_id"
		actualBuild, err := Push(tmpFile, appID, c)

		assert.NoError(t, err)
		assert.Equal(t, expectedBuild, actualBuild)
	})

	t.Run("can return error on AppCreate mutation", func(t *testing.T) {
		buildPushFailedResponse := `{"errors":[{"message":"Something went wrong","path":["buildPush"]}],"data":null}`
		c := createTestGqlClient(buildPushFailedResponse)

		appID := "app_id"
		actualBuild, err := Push(tmpFile, appID, c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Something went wrong")
		assert.Equal(t, BuildConfiguration{}, actualBuild)
	})
}

func createTestGqlClient(response string) *gqlclient.Client {
	h := http.Header{}
	h.Add("Content-Type", "application/json")
	ts := test.TestTransport{
		WithResponse: &http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts})
}
