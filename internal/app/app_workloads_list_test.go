package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/test"
)

func TestAppWorkloadsList(t *testing.T) {
	t.Run("given successful response it returns expected app workloads", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appWorkloadsRespBody := `
				{
					"data": {
						"appWorkloads": [
							{
								"organization": {
									"slug": "test-organization-slug"
								},
								"startedAt": "2024-01-01T13:00:00.000Z",
								"status": "RUNNING",
								"logs": {
									"edges": [
										{"timestamp": "2024-01-01T13:00:01.000Z", "text": "log entry 1"},
										{"timestamp": "2024-01-01T13:00:02.000Z", "text": "log entry 2"},
										{"timestamp": "2024-01-01T13:00:03.000Z", "text": "log entry 3"}
									]
								},
								"cpuUsage": {
									"current": 5.00
								},
								"memoryUsageMB": {
									"current": 123.5
								}
							},
							{
								"subscription": {
									"id": "test-subscription-id",
									"inboundOrganization": {
										"slug": "test-subscribing-organization-slug"
									}
								},
								"startedAt": "2024-02-02T14:00:00.000Z",
								"status": "RUNNING",
								"logs": {
									"edges": [
										{"timestamp": "2024-02-02T14:00:01.000Z", "text": "log entry 1"},
										{"timestamp": "2024-02-02T14:00:02.000Z", "text": "log entry 2"},
										{"timestamp": "2024-02-02T14:00:03.000Z", "text": "log entry 3"}
									]
								},
								"cpuUsage": {
									"current": 5.00,
									"limit": 10.0
								},
								"memoryUsageMB": {
									"current": 123.5,
									"limit": 1024.0
								}
							}
						]
					}
				}
			`
		resp := test.JSONResponse(appWorkloadsRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := ListAppWorkloadsInput{AppID: "test-app-id"}
		actual, err := s.ListAppWorkloads(context.TODO(), input)

		expected := []AppWorkload{
			{
				OrganizationSlug: "test-organization-slug",
				Subscription:     nil,
				StartedAt:        time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC),
				Status:           "RUNNING",
				LogEntries: []AppDeployLogEntry{
					{Timestamp: time.Date(2024, time.January, 1, 13, 0, 1, 0, time.UTC), Text: "log entry 1"},
					{Timestamp: time.Date(2024, time.January, 1, 13, 0, 2, 0, time.UTC), Text: "log entry 2"},
					{Timestamp: time.Date(2024, time.January, 1, 13, 0, 3, 0, time.UTC), Text: "log entry 3"},
				},
				CPUUsage: AppWorkloadResourceUsage{
					Current: 5.0,
					Limit:   nil,
				},
				MemoryUsageMB: AppWorkloadResourceUsage{
					Current: 123.5,
					Limit:   nil,
				},
			},
			{
				OrganizationSlug: "",
				Subscription: &AppWorkloadSubscription{
					OrganizationSlug: "test-subscribing-organization-slug",
					SubscriptionUUID: "test-subscription-id",
				},
				StartedAt: time.Date(2024, time.February, 2, 14, 0, 0, 0, time.UTC),
				Status:    "RUNNING",
				LogEntries: []AppDeployLogEntry{
					{Timestamp: time.Date(2024, time.February, 2, 14, 0, 1, 0, time.UTC), Text: "log entry 1"},
					{Timestamp: time.Date(2024, time.February, 2, 14, 0, 2, 0, time.UTC), Text: "log entry 2"},
					{Timestamp: time.Date(2024, time.February, 2, 14, 0, 3, 0, time.UTC), Text: "log entry 3"},
				},
				CPUUsage: AppWorkloadResourceUsage{
					Current: 5.0,
					Limit:   ref(10.0),
				},
				MemoryUsageMB: AppWorkloadResourceUsage{
					Current: 123.5,
					Limit:   ref(1024.0),
				},
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("given graphql error it returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appWorkloadsRespBody := `
				{
					"errors": [{
						"message": "test error message",
						"location": [{"line": 1, "column": 1}],
						"path": ["appWorkloads"]
					}]
				}
			`
		resp := test.JSONResponse(appWorkloadsRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := ListAppWorkloadsInput{AppID: "test-app-id"}
		actual, err := s.ListAppWorkloads(context.TODO(), input)

		assert.Nil(t, actual)
		assert.ErrorContains(t, err, "test error message")
	})

	t.Run("given access denied error then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appWorkloadsRespBody := `
				{
					"errors": [{
						"message": "access denied",
						"location": [{"line": 1, "column": 1}],
						"path": ["appWorkloads"]
					}]
				}
			`
		resp := test.JSONResponse(appWorkloadsRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := ListAppWorkloadsInput{AppID: "test-app-id"}
		actual, err := s.ListAppWorkloads(context.TODO(), input)

		assert.Nil(t, actual)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})
}
