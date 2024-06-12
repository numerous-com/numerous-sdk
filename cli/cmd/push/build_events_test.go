package push

import (
	"bytes"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

type fakeSubscriptionEvent struct {
	message []byte
	err     error
}

type fakeSubscriptionClient struct {
	events []fakeSubscriptionEvent
}

func (sc *fakeSubscriptionClient) Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error) {
	for _, ev := range sc.events {
		if err := handler(ev.message, ev.err); err != nil {
			return "", err
		}
	}

	return "", nil
}

func TestBuildEvents(t *testing.T) {
	client := fakeSubscriptionClient{
		events: []fakeSubscriptionEvent{
			{
				message: []byte(`{"buildEvents": {"__typename": "BuildEventInfo", "result": "{\"stream\": \"Some message\"}"}}`),
			},
		},
	}

	buf := bytes.NewBuffer(nil)
	err := buildEventSubscription(&client, buf, "some build ID", "", true)

	assert.NoError(t, err)
	assert.Equal(t, "Some message", buf.String())
}
