package status

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/app"
)

var errTest = errors.New("test error")

func TestStatus(t *testing.T) {
	ctx := context.Background()
	appID := "test-app-id"
	appSlug := "test-app-slug"
	orgSlug := "test-org-slug"
	appDir := t.TempDir()
	input := statusInput{appSlug: appSlug, orgSlug: orgSlug, appDir: appDir}
	readAppinput := app.ReadAppInput{OrganizationSlug: orgSlug, AppSlug: appSlug}
	readAppOutput := app.ReadAppOutput{AppID: appID}
	listAppWorkloadsInput := app.ListAppWorkloadsInput{AppID: appID}
	appWorkloads := []app.AppWorkload{
		{
			OrganizationSlug: "test-organization-slug",
			StartedAt:        time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC),
			Status:           "RUNNING",
			LogEntries: []app.AppDeployLogEntry{
				{Timestamp: time.Date(2024, time.January, 1, 13, 0, 1, 0, time.UTC), Text: "log entry 1"},
				{Timestamp: time.Date(2024, time.January, 1, 13, 0, 2, 0, time.UTC), Text: "log entry 2"},
				{Timestamp: time.Date(2024, time.January, 1, 13, 0, 3, 0, time.UTC), Text: "log entry 3"},
			},
			CPUUsage:      app.AppWorkloadResourceUsage{Current: 10.0},
			MemoryUsageMB: app.AppWorkloadResourceUsage{Current: 20.0},
		},
		{
			Subscription: &app.AppWorkloadSubscription{OrganizationSlug: "test-subscribing-organization-slug", SubscriptionUUID: "test-subscription-id"},
			StartedAt:    time.Date(2024, time.February, 2, 14, 0, 0, 0, time.UTC),
			Status:       "RUNNING",
			LogEntries: []app.AppDeployLogEntry{
				{Timestamp: time.Date(2024, time.February, 2, 14, 0, 1, 0, time.UTC), Text: "log entry 1"},
				{Timestamp: time.Date(2024, time.February, 2, 14, 0, 2, 0, time.UTC), Text: "log entry 2"},
				{Timestamp: time.Date(2024, time.February, 2, 14, 0, 3, 0, time.UTC), Text: "log entry 3"},
			},
			CPUUsage:      app.AppWorkloadResourceUsage{Current: 10.0, Limit: ref(20.0)},
			MemoryUsageMB: app.AppWorkloadResourceUsage{Current: 20.0, Limit: ref(40.0)},
		},
	}

	t.Run("makes expected app read call", func(t *testing.T) {
		mockApps := &mockAppReaderWorkloadLister{}
		mockApps.On("ReadApp", mock.Anything, mock.Anything).Return(readAppOutput, nil)
		mockApps.On("ListAppWorkloads", mock.Anything, mock.Anything).Return(appWorkloads, nil)

		err := status(ctx, mockApps, input)

		assert.NoError(t, err)
		mockApps.AssertNumberOfCalls(t, "ReadApp", 1)
		mockApps.AssertCalled(t, "ReadApp", ctx, readAppinput)
	})

	t.Run("makes expected app workload list call", func(t *testing.T) {
		mockApps := &mockAppReaderWorkloadLister{}
		mockApps.On("ReadApp", mock.Anything, mock.Anything).Return(readAppOutput, nil)
		mockApps.On("ListAppWorkloads", mock.Anything, mock.Anything).Return(appWorkloads, nil)

		err := status(ctx, mockApps, input)

		assert.NoError(t, err)
		mockApps.AssertNumberOfCalls(t, "ListAppWorkloads", 1)
		mockApps.AssertCalled(t, "ListAppWorkloads", ctx, listAppWorkloadsInput)
	})

	t.Run("given app read error it is returned and app workload list is not called", func(t *testing.T) {
		mockApps := &mockAppReaderWorkloadLister{}
		mockApps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{}, errTest)

		err := status(ctx, mockApps, input)

		assert.ErrorIs(t, err, errTest)
		mockApps.AssertNotCalled(t, "ListAppWorkloads")
	})

	t.Run("given error from app workload list call it is returned", func(t *testing.T) {
		mockApps := &mockAppReaderWorkloadLister{}
		mockApps.On("ReadApp", mock.Anything, mock.Anything).Return(readAppOutput, nil)
		mockApps.On("ListAppWorkloads", mock.Anything, mock.Anything).Return(([]app.AppWorkload)(nil), errTest)

		err := status(ctx, mockApps, input)

		assert.ErrorIs(t, err, errTest)
	})
}

func TestHumanDuration(t *testing.T) {
	type testCase struct {
		name     string
		duration time.Duration
		expected string
	}

	for _, tc := range []testCase{
		{
			name:     "seconds only",
			duration: 5 * time.Second,
			expected: "5 seconds",
		},
		{
			name:     "seconds are rounded",
			duration: 7*time.Second + 10*time.Millisecond + 20*time.Microsecond,
			expected: "7 seconds",
		},
		{
			name:     "minutes and seconds",
			duration: 123 * time.Second,
			expected: "2 minutes and 3 seconds",
		},
		{
			name:     "hours and minutes",
			duration: 123 * time.Minute,
			expected: "2 hours and 3 minutes",
		},
		{
			name:     "days and hours",
			duration: 50 * time.Hour,
			expected: "2 days and 2 hours",
		},
	} {
		actual := humanizeDuration(tc.duration)
		assert.Equal(t, tc.expected, actual)
	}
}

func ref[T any](v T) *T {
	return &v
}
