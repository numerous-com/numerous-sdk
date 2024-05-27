package test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/mock"
)

type validateDoer struct {
	t     *testing.T
	inner graphql.Doer
}

func (d *validateDoer) Do(r *http.Request) (*http.Response, error) {
	resp, err := d.inner.Do(r)
	if err != nil {
		return resp, err
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(data))
	assertQuery(d.t, r)

	// recreate body
	resp.Body = io.NopCloser(bytes.NewReader(data))

	return resp, err
}

func CreateTestGQLClient(t *testing.T, doer *MockDoer) *graphql.Client {
	t.Helper()

	validateDoer := validateDoer{t, doer}

	return graphql.NewClient("http://url", &validateDoer)
}

var _ graphql.Doer = &MockDoer{}

type MockDoer struct {
	mock.Mock
}

func (m *MockDoer) Do(r *http.Request) (*http.Response, error) {
	args := m.Called(r)

	return args.Get(0).(*http.Response), args.Error(1)
}

func JSONResponse(json string) *http.Response {
	return &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(json))),
	}
}
