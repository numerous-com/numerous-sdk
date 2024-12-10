package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var dummyData = []byte("some data")

func dummyReader() io.Reader {
	return bytes.NewBuffer(dummyData)
}

func TestUploadAppSource(t *testing.T) {
	testError := errors.New("some test error")

	t.Run("given http client error then it returns error", func(t *testing.T) {
		doer := test.MockDoer{}
		var nilResp *http.Response
		doer.On("Do", mock.Anything).Return(nilResp, testError)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", UploadArchive{Reader: dummyReader(), Size: int64(len(dummyData))})

		assert.ErrorIs(t, err, testError)
	})

	t.Run("given non-OK http status then it returns error", func(t *testing.T) {
		doer := test.MockDoer{}
		responseBody := []byte("error response body")
		resp := http.Response{Status: "Not OK", StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBuffer(responseBody))}
		doer.On("Do", mock.Anything).Return(&resp, nil)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", UploadArchive{Reader: dummyReader(), Size: int64(len(dummyData))})

		expected := AppSourceUploadError{
			HTTPStatusCode: http.StatusBadRequest,
			HTTPStatus:     "Not OK",
			UploadURL:      "http://some-upload-url",
			ResponseBody:   responseBody,
		}
		actual := &AppSourceUploadError{}
		if assert.ErrorAs(t, err, &actual) {
			assert.Equal(t, expected, *actual)
		}
	})

	t.Run("given invalid upload URL then it returns error", func(t *testing.T) {
		s := Service{uploadDoer: &test.MockDoer{}}

		err := s.UploadAppSource("://invalid-url", UploadArchive{Reader: dummyReader(), Size: int64(len(dummyData))})

		assert.Error(t, err)
	})

	t.Run("given successful request then it returns no error", func(t *testing.T) {
		doer := test.MockDoer{}
		resp := http.Response{Status: "OK", StatusCode: http.StatusOK}
		doer.On("Do", mock.Anything).Return(&resp, nil)
		s := Service{uploadDoer: &doer}

		err := s.UploadAppSource("http://some-upload-url", UploadArchive{Reader: bytes.NewReader([]byte("")), Size: 0})

		assert.NoError(t, err)
	})

	t.Run("it sends expected content-length header", func(t *testing.T) {
		doer := test.MockDoer{}
		resp := http.Response{Status: "OK", StatusCode: http.StatusOK}
		doer.On("Do", mock.Anything).Return(&resp, nil)
		s := Service{uploadDoer: &doer}
		data := []byte("some data")
		err := s.UploadAppSource("http://some-upload-url", UploadArchive{Reader: bytes.NewReader(data), Size: int64(len(data))})

		assert.NoError(t, err)
		doer.AssertCalled(t, "Do", mock.MatchedBy(func(r *http.Request) bool {
			return r.ContentLength == 9
		}))
	})
}
