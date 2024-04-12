package test

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// HTTPTransport implements a http.RoundTripper for testing purposes only.
type TestTransport struct {
	WithResponse *http.Response
	WithError    error
	Requests     []*http.Request
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Requests = append(t.Requests, req)

	return t.WithResponse, t.WithError
}

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	return args.Get(0).(*http.Response), args.Error(1)
}
