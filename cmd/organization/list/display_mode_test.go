package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayMode(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		type testCase struct {
			value    string
			expected DisplayMode
		}

		for _, tc := range []testCase{
			{
				value:    "list",
				expected: DisplayModeList,
			},
			{
				value:    "LIST",
				expected: DisplayModeList,
			},
			{
				value:    "table",
				expected: DisplayModeTable,
			},
			{
				value:    "TABLE",
				expected: DisplayModeTable,
			},
		} {
			t.Run(tc.value, func(t *testing.T) {
				var d DisplayMode

				err := d.Set(tc.value)

				assert.NoError(t, err)
				assert.Equal(t, tc.expected, d)
			})
		}

		t.Run("error", func(t *testing.T) {
			var d DisplayMode

			err := d.Set("invalid display mode")

			assert.ErrorIs(t, err, errInvalidDisplayMode)
		})
	})

	t.Run("String", func(t *testing.T) {
		type testCase struct {
			value    DisplayMode
			expected string
		}

		for _, tc := range []testCase{
			{
				expected: "list",
				value:    DisplayModeList,
			},
			{
				expected: "table",
				value:    DisplayModeTable,
			},
		} {
			t.Run(tc.expected, func(t *testing.T) {
				actual := tc.value.String()

				assert.Equal(t, tc.expected, actual)
			})
		}
	})

	t.Run("Type", func(t *testing.T) {
		var d DisplayMode

		actual := d.Type()

		assert.Equal(t, "Display mode", actual)
	})
}
