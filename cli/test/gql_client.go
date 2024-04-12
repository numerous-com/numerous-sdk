package test

import (
	"bytes"
	"io"
	"net/http"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/stretchr/testify/mock"
)

func CreateTestGqlClient(response string) *gqlclient.Client {
	h := http.Header{}
	h.Add("Content-Type", "application/json")

	ts := TestTransport{
		WithResponse: &http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts})
}

func CreateMockGqlClient(responses ...string) (*gqlclient.Client, *MockTransport) {
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
