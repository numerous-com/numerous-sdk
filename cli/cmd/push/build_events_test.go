package push

import (
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
				message: []byte("{}"),
			},
		},
	}

	err := buildEventSubscription(&client, "some build ID", "", false)

	assert.NoError(t, err)
}
