package test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/stretchr/testify/mock"

	_ "embed"
)

func CreateTestGqlClient(t *testing.T, response string) *gqlclient.Client {
	t.Helper()

	h := http.Header{}
	h.Add("Content-Type", "application/json")

	ts := TestTransport{
		WithResponse: &http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
		Handler: func(r *http.Request) *struct {
			Response *http.Response
			Error    error
		} {
			assertGraphQLRequest(t, r)
			return nil
		},
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts})
}

func CreateMockGQLClient(responses ...string) (*gqlclient.Client, *MockTransport) {
	ts := MockTransport{}

	for _, response := range responses {
		AddResponseToMockGqlClient(response, &ts)
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts}), &ts
}

func AddResponseToMockGqlClient(response string, ts *MockTransport) {
	h := http.Header{}
	h.Add("Content-Type", "application/json")

	ts.On("RoundTrip", mock.Anything).Once().Return(
		&http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
		nil,
	)
}
