package logging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	testcases := []struct {
		value    string
		expected Level
	}{
		{value: "debug", expected: LevelDebug},
		{value: "info", expected: LevelInfo},
		{value: "warning", expected: LevelWarning},
		{value: "error", expected: LevelError},
		{value: "DEBUG", expected: LevelDebug},
		{value: "INFO", expected: LevelInfo},
		{value: "WARNING", expected: LevelWarning},
		{value: "ERROR", expected: LevelError},
	}

	for _, testcase := range testcases {
		t.Run(testcase.value, func(t *testing.T) {
			var l Level
			err := l.Set(testcase.value)

			require.NoError(t, err)
			assert.Equal(t, testcase.expected, l)
		})
	}

	t.Run("returns error on invalid value", func(t *testing.T) {
		var l Level

		err := l.Set("some other value")
		require.ErrorIs(t, err, ErrInvalidLogLevel)
	})
}

func TestString(t *testing.T) {
	testcases := []struct {
		level    Level
		expected string
	}{
		{expected: "debug", level: LevelDebug},
		{expected: "info", level: LevelInfo},
		{expected: "warning", level: LevelWarning},
		{expected: "error", level: LevelError},
	}

	for _, testcase := range testcases {
		t.Run(testcase.expected, func(t *testing.T) {
			actual := testcase.level.String()
			assert.Equal(t, testcase.expected, actual)
		})
	}
}

func TestToSlogLevel(t *testing.T) {
	testcases := []struct {
		level    Level
		expected slog.Level
	}{
		{level: LevelDebug, expected: slog.LevelDebug},
		{level: LevelInfo, expected: slog.LevelInfo},
		{level: LevelWarning, expected: slog.LevelWarn},
		{level: LevelError, expected: slog.LevelError},
	}

	for _, testcase := range testcases {
		t.Run(testcase.level.String(), func(t *testing.T) {
			slogLevel := testcase.level.ToSlogLevel()
			assert.Equal(t, testcase.expected, slogLevel)
		})
	}
}

func TestType(t *testing.T) {
	levels := []Level{LevelDebug, LevelInfo, LevelWarning, LevelError}
	for _, level := range levels {
		assert.Equal(t, "Log level", level.Type())
	}
}
