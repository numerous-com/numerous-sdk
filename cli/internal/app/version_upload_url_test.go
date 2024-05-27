package app

import (
	"context"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAppVersionUploadURL(t *testing.T) {
	t.Run("get app version upload url returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)

		respBody := `
			{
				"data": {
					"appVersionUploadURL": {
						"url": "https://upload-url.com/url-path"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := GetAppVersionUploadURLInput{AppVersionID: "some-app-version-id"}
		output, err := GetAppVersionUploadURL(context.TODO(), c, input)

		expected := GetAppVersionUploadURLOutput{UploadURL: "https://upload-url.com/url-path"}
		if assert.NoError(t, err) {
			assert.Equal(t, expected, output)
		}
	})

	t.Run("get app version upload url returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)

		respBody := `
			{
				"errors": [{
					"message": "expected error message",
					"location": [{"line": 1, "column": 1}],
					"path": ["appVersionUploadURL"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := GetAppVersionUploadURLInput{AppVersionID: "some-app-version-id"}
		output, err := GetAppVersionUploadURL(context.TODO(), c, input)

		expected := GetAppVersionUploadURLOutput{}
		if assert.ErrorContains(t, err, "expected error message") {
			assert.Equal(t, expected, output)
		}
	})
}
