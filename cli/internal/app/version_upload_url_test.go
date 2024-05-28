package app

import (
	"context"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAppVersionUploadURL(t *testing.T) {
	t.Run("returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

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

		input := AppVersionUploadURLInput{AppVersionID: "some-app-version-id"}
		output, err := s.AppVersionUploadURL(context.TODO(), input)

		expected := AppVersionUploadURLOutput{UploadURL: "https://upload-url.com/url-path"}
		if assert.NoError(t, err) {
			assert.Equal(t, expected, output)
		}
	})

	t.Run("returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

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

		input := AppVersionUploadURLInput{AppVersionID: "some-app-version-id"}
		output, err := s.AppVersionUploadURL(context.TODO(), input)

		expected := AppVersionUploadURLOutput{}
		if assert.ErrorContains(t, err, "expected error message") {
			assert.Equal(t, expected, output)
		}
	})
}
