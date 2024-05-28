package app

import (
	"context"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestDeployEvents(t *testing.T) {
	t.Run("returns expected output", func(t *testing.T) {
		ch := make(chan test.SubMessage)
		c := test.CreateTestSubscriptionClient(t, ch)
		s := New(nil, c, nil)

		actualEvents := []DeployEvent{}
		input := DeployEventsInput{
			Handler: func(ev DeployEvent) bool {
				actualEvents = append(actualEvents, ev)
				return true
			},
			DeploymentVersionID: "some-id",
		}
		err := s.DeployEvents(context.TODO(), input)
		ch <- test.SubMessage{Msg: `{"appDeployLogs": {"message": "message 1"}}`}
		ch <- test.SubMessage{Msg: `{"appDeployLogs": {"message": "message 2"}}`}
		ch <- test.SubMessage{Msg: `{"appDeployLogs": {"message": "message 3"}}`}
		close(ch)

		expected := []DeployEvent{
			{Message: "message 1"},
			{Message: "message 2"},
			{Message: "message 3"},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, actualEvents)
	})
}
