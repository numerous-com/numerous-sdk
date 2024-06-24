package app

import (
	"context"
	"testing"
	"time"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestDeployEvents(t *testing.T) {
	testCases := []struct {
		name     string
		sms      []test.SubMessage
		expected []DeployEvent
	}{
		{
			name: "returns expected build messages",
			sms: []test.SubMessage{
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildMessageEvent", "message": "message 1"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildMessageEvent", "message": "message 2"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildMessageEvent", "message": "message 3"}}`},
			},
			expected: []DeployEvent{
				{Typename: "AppBuildMessageEvent", BuildMessage: AppBuildMessageEvent{Message: "message 1"}},
				{Typename: "AppBuildMessageEvent", BuildMessage: AppBuildMessageEvent{Message: "message 2"}},
				{Typename: "AppBuildMessageEvent", BuildMessage: AppBuildMessageEvent{Message: "message 3"}},
			},
		},
		{
			name: "returns expected deploy status events",
			sms: []test.SubMessage{
				{Msg: `{"appDeployEvents": {"__typename": "AppDeploymentStatusEvent", "status": "PENDING"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppDeploymentStatusEvent", "status": "RUNNING"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppDeploymentStatusEvent", "status": "STOPPED"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppDeploymentStatusEvent", "status": "ERROR"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppDeploymentStatusEvent", "status": "UNKNOWN"}}`},
			},
			expected: []DeployEvent{
				{Typename: "AppDeploymentStatusEvent", DeploymentStatus: AppDeploymentStatusEvent{Status: "PENDING"}},
				{Typename: "AppDeploymentStatusEvent", DeploymentStatus: AppDeploymentStatusEvent{Status: "RUNNING"}},
				{Typename: "AppDeploymentStatusEvent", DeploymentStatus: AppDeploymentStatusEvent{Status: "STOPPED"}},
				{Typename: "AppDeploymentStatusEvent", DeploymentStatus: AppDeploymentStatusEvent{Status: "ERROR"}},
				{Typename: "AppDeploymentStatusEvent", DeploymentStatus: AppDeploymentStatusEvent{Status: "UNKNOWN"}},
			},
		},
		{
			name: "returns expected build errors",
			sms: []test.SubMessage{
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildErrorEvent", "message": "error 1"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildErrorEvent", "message": "error 2"}}`},
				{Msg: `{"appDeployEvents": {"__typename": "AppBuildErrorEvent", "message": "error 3"}}`},
			},
			expected: []DeployEvent{
				{Typename: "AppBuildErrorEvent", BuildError: AppBuildErrorEvent{Message: "error 1"}},
				{Typename: "AppBuildErrorEvent", BuildError: AppBuildErrorEvent{Message: "error 2"}},
				{Typename: "AppBuildErrorEvent", BuildError: AppBuildErrorEvent{Message: "error 3"}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan test.SubMessage, 10)
			c := test.CreateTestSubscriptionClient(t, ch)
			s := New(nil, c, nil)

			actual := []DeployEvent{}
			input := DeployEventsInput{
				Handler: func(ev DeployEvent) error {
					actual = append(actual, ev)
					return nil
				},
				DeploymentVersionID: "some-id",
			}
			go func() {
				defer close(ch)
				time.Sleep(time.Millisecond * 10)
				for _, sm := range tc.sms {
					ch <- sm
				}
			}()
			err := s.DeployEvents(context.TODO(), input)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
