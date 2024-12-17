package status

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanizeDuration(t *testing.T) {
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
