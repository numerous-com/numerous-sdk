package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadAppSource(t *testing.T) {
	testError := errors.New("some test error")

	t.Run("given http client error then it returns error", func(t *testing.T) {
		doer := test.MockDoer{}
		var nilResp *http.Response
		doer.On("Do", mock.Anything, mock.Anything).Return(nilResp, testError)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", nil)

		assert.ErrorIs(t, err, testError)
	})

	t.Run("given non-OK http status then it returns error", func(t *testing.T) {
		doer := test.MockDoer{}
		resp := http.Response{Status: "Not OK", StatusCode: http.StatusBadRequest}
		doer.On("Do", mock.Anything, mock.Anything).Return(&resp, nil)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", nil)

		assert.ErrorIs(t, err, ErrAppSourceUpload)
	})

	t.Run("given invalid upload URL then it returns error", func(t *testing.T) {
		s := Service{uploadDoer: &test.MockDoer{}}

		err := s.UploadAppSource("://invalid-url", nil)

		assert.Error(t, err)
	})

	t.Run("given successful request then it returns no error", func(t *testing.T) {
		doer := test.MockDoer{}
		resp := http.Response{Status: "OK", StatusCode: http.StatusOK}
		doer.On("Do", mock.Anything, mock.Anything).Return(&resp, nil)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", io.NopCloser(bytes.NewReader([]byte(""))))

		assert.NoError(t, err)
	})
}
