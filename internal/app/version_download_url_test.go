package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAppVersionDownloadURL(t *testing.T) {
	t.Run("returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"appVersionDownloadURL": {
						"url": "https://download-url.com/url-path"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := AppVersionDownloadURLInput{AppVersionID: "some-app-version-id"}
		output, err := s.AppVersionDownloadURL(context.TODO(), input)

		expected := AppVersionDownloadURLOutput{DownloadURL: "https://download-url.com/url-path"}
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
					"path": ["appVersionDownloadURL"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := AppVersionDownloadURLInput{AppVersionID: "some-app-version-id"}
		output, err := s.AppVersionDownloadURL(context.TODO(), input)

		expected := AppVersionDownloadURLOutput{}
		if assert.ErrorContains(t, err, "expected error message") {
			assert.Equal(t, expected, output)
		}
	})
}
