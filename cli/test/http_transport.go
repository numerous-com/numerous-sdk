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
	Handler      func(*http.Request) *struct {
		Response *http.Response
		Error    error
	}
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Requests = append(t.Requests, req)

	if t.Handler != nil {
		if res := t.Handler(req); res != nil {
			return res.Response, res.Error
		}
	}

	return t.WithResponse, t.WithError
}

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	return args.Get(0).(*http.Response), args.Error(1)
}
