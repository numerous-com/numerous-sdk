package app

import (
	"sync"
	"testing"
	"time"

	"numerous.com/cli/cmd/app/appident"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestAppDeployLogs(t *testing.T) {
	clientCh := make(chan test.SubMessage, 10)
	c := test.CreateTestSubscriptionClient(t, clientCh)
	s := New(nil, c, nil)

	ch, err := s.AppDeployLogs(appident.AppIdentifier{OrganizationSlug: "organization-slug", Name: "app-name"})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	actual := []AppDeployLogEntry{}
	go func() {
		defer wg.Done()

		for {
			select {
			case <-time.After(time.Second):
				return
			case e, ok := <-ch:
				if !ok {
					return
				}
				actual = append(actual, e)
			}
		}
	}()
	clientCh <- test.SubMessage{Msg: `{"appDeployLogs": {"timestamp": "2024-11-11T11:11:11Z", "text": "message 1"}}`}
	clientCh <- test.SubMessage{Msg: `{"appDeployLogs": {"timestamp": "2024-11-11T11:11:22Z", "text": "message 2"}}`}
	clientCh <- test.SubMessage{Msg: `{"appDeployLogs": {"timestamp": "2024-11-11T11:11:33Z", "text": "message 3"}}`}
	close(clientCh)
	wg.Wait()

	expected := []AppDeployLogEntry{
		{Timestamp: time.Date(2024, time.November, 11, 11, 11, 11, 0, time.UTC), Text: "message 1"},
		{Timestamp: time.Date(2024, time.November, 11, 11, 11, 22, 0, time.UTC), Text: "message 2"},
		{Timestamp: time.Date(2024, time.November, 11, 11, 11, 33, 0, time.UTC), Text: "message 3"},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
