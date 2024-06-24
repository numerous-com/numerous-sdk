package test

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
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
	assertGraphQLRequest(d.t, r)

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

type validatingSubscriptionClient struct {
	t       *testing.T
	handler func(message []byte, err error) error
	ch      chan SubMessage
}

func (c *validatingSubscriptionClient) Run() error {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for ev := range c.ch {
			c.handler([]byte(ev.Msg), ev.Err) //nolint:errcheck
		}
	}()

	select {
	case <-done:
		break
	case <-time.After(time.Second):
		assert.Fail(c.t, "timed out waiting for subscription to close")
	}

	return nil
}

func (c *validatingSubscriptionClient) Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error) {
	assertSubscription(c.t, v, variables)
	c.handler = handler

	return "subID", nil
}

func (c *validatingSubscriptionClient) Close() error { return nil }

type SubMessage struct {
	Msg string
	Err error
}

func CreateTestSubscriptionClient(t *testing.T, ch chan SubMessage) *validatingSubscriptionClient {
	t.Helper()

	return &validatingSubscriptionClient{t, nil, ch}
}
