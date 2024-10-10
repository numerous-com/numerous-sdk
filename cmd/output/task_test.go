package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	t.Run("it generates expected lines", func(t *testing.T) {
		type testCase struct {
			msg               string
			icon              string
			lineWidth         int
			expectedLineWidth int
			expected          string
		}
		for _, tc := range []testCase{
			{
				msg:               "message that is too long",
				lineWidth:         20,
				expectedLineWidth: 20,
				icon:              "⏰",
				expected:          "⏰ message th...",
			},

			{
				msg:               "a message that is not too long",
				lineWidth:         45,
				expectedLineWidth: 45,
				icon:              "⚠",
				expected:          "⚠ a message that is not too long........",
			},
			{
				msg:               "a message with a longer icon is interpreted as 1 length",
				lineWidth:         70,
				expectedLineWidth: 70,
				icon:              "<long icon>",
				expected:          "<long icon> a message with a longer icon is interpreted as 1 length........",
			},
			{
				msg:               "even without message the line is too narrow cannot be trimmed",
				lineWidth:         1,
				expectedLineWidth: 10,
				icon:              "<icon>",
				expected:          "<icon> ...",
			},
		} {
			t.Run(tc.msg, func(t *testing.T) {
				task := Task{
					msg:       tc.msg,
					lineWidth: func() int { return tc.lineWidth },
				}

				actual := task.line(tc.icon)

				actual = strings.Replace(actual, AnsiFaint, "", 1)
				actual = strings.Replace(actual, AnsiReset, "", 1)
				extraIconLength := len(tc.icon) - 1
				assert.Equal(t, tc.expected, actual)
				assert.Len(t, actual+"Error", tc.expectedLineWidth+extraIconLength)
			})
		}
	})
}
