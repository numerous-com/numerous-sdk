package status

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSince(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Run("sets value", func(t *testing.T) {
			var actual Since

			err := actual.Set("2024-01-01T10:11:12Z")

			assert.NoError(t, err)
			expected := time.Date(2024, time.January, 1, 10, 11, 12, 0, time.UTC)
			assert.Equal(t, Since(expected), actual)
		})

		t.Run("returns error", func(t *testing.T) {
			var actual Since

			err := actual.Set("invalid-since-value")

			assert.ErrorIs(t, err, errParseSince)
		})
	})
}

func TestParseSince(t *testing.T) {
	now := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	type testCase struct {
		value    string
		expected time.Time
	}

	for _, tc := range []testCase{
		{value: "1h", expected: now.Add(-time.Hour)},
		{value: "123h", expected: now.Add(-123 * time.Hour)},
		{value: "5d", expected: now.Add(-5 * 24 * time.Hour)},
		{value: "1000d", expected: now.Add(-1000 * 24 * time.Hour)},
		{value: "3m", expected: now.Add(-3 * time.Minute)},
		{value: "120m", expected: now.Add(-120 * time.Minute)},
		{value: "2s", expected: now.Add(-2 * time.Second)},
		{value: "8600s", expected: now.Add(-8600 * time.Second)},
		{value: "2024-01-01T10:11:12Z", expected: time.Date(2024, time.January, 1, 10, 11, 12, 0, time.UTC)},
		{value: "2024-01-01T10:11:12+02:00", expected: time.Date(2024, time.January, 1, 10, 11, 12, 0, time.FixedZone("", int((2*time.Hour).Seconds())))},
		{value: "2024-01-01", expected: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.value, func(t *testing.T) {
			actual, err := parseSince(tc.value, now)

			assert.NoError(t, err)
			if assert.NotNil(t, actual) {
				assert.Equal(t, Since(tc.expected), *actual)
			}
		})
	}
}
