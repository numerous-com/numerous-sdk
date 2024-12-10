package output

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

var errTest = errors.New("test error")

var _ terminal = &stubTerminal{}

type stubTerminal struct {
	width         int
	isNotTerminal bool
	buf           *bytes.Buffer
	getSizeErr    error
}

func (s *stubTerminal) Writer() io.Writer {
	return s.buf
}

func (s *stubTerminal) GetSize() (int, int, error) {
	return s.width, 0, s.getSizeErr
}

func (s *stubTerminal) IsTerminal() bool {
	return !s.isNotTerminal
}

func TestTask(t *testing.T) {
	type testCase struct {
		msg           string
		termWidth     int
		expected      string
		getSizeErr    error
		isNotTerminal bool
	}

	newStubTerminal := func(tc testCase, buf *bytes.Buffer) *stubTerminal {
		return &stubTerminal{isNotTerminal: tc.isNotTerminal, width: tc.termWidth, buf: buf, getSizeErr: tc.getSizeErr}
	}

	testCases := []testCase{
		{
			msg:       "message that is too long is trimmed",
			termWidth: 20,
			expected:  " message th" + AnsiFaint + "..." + AnsiReset,
		},

		{
			msg:       "a message that is not too long is not trimmed",
			termWidth: 60,
			expected:  " a message that is not too long is not trimmed" + AnsiFaint + "........" + AnsiReset,
		},
		{
			msg:       "even without message the line is too narrow cannot be trimmed",
			termWidth: 1,
			expected:  " " + AnsiFaint + "..." + AnsiReset,
		},
		{
			msg:       "no matter the terminal width the maximum width is not exceeded",
			termWidth: 1000,
			expected:  " no matter the terminal width the maximum width is not exceeded" + AnsiFaint + "..................................................." + AnsiReset,
		},
		{
			msg:        "uses fallback width with GetSize error",
			expected:   " uses fallback width with GetSize error" + AnsiFaint + "..............." + AnsiReset,
			getSizeErr: errTest,
		},
		{
			msg:           "uses fallback width when not terminal error",
			expected:      " uses fallback width when not terminal error" + AnsiFaint + ".........." + AnsiReset,
			isNotTerminal: true,
		},
	}
	t.Run("StartTaskWithTerminal", func(t *testing.T) {
		t.Run("start task writes start line with hourglass", func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.msg, func(t *testing.T) {
					buf := bytes.NewBuffer(nil)
					term := newStubTerminal(tc, buf)

					StartTaskWithTerminal(tc.msg, term)

					actual := buf.String()
					expected := hourglassIcon + tc.expected
					assert.Equal(t, expected, actual)
				})
			}
		})
	})

	t.Run("Done", func(t *testing.T) {
		t.Run("writes expected updated and terminated line", func(t *testing.T) {
			greenOK := AnsiGreen + "OK" + AnsiReset
			for _, tc := range testCases {
				t.Run(tc.msg, func(t *testing.T) {
					buf := bytes.NewBuffer(nil)
					term := newStubTerminal(tc, buf)

					task := StartTaskWithTerminal(tc.msg, term)
					buf.Truncate(0) // remove existing content from buffer
					task.Done()

					actual := buf.String()
					expected := "\r" + checkmarkIcon + tc.expected + greenOK + "\n"
					assert.Equal(t, expected, actual)
				})
			}
		})
	})

	t.Run("Error", func(t *testing.T) {
		t.Run("writes expected updated and terminated line", func(t *testing.T) {
			redError := AnsiRed + "Error" + AnsiReset
			for _, tc := range testCases {
				t.Run(tc.msg, func(t *testing.T) {
					buf := bytes.NewBuffer(nil)
					term := newStubTerminal(tc, buf)

					task := StartTaskWithTerminal(tc.msg, term)
					buf.Truncate(0) // remove existing content from buffer
					task.Error()

					actual := buf.String()
					expected := "\r" + errorcross + tc.expected + redError + "\n"
					assert.Equal(t, expected, actual)
				})
			}
		})
	})

	t.Run("AddLine", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		term := &stubTerminal{buf: buf, width: 25}

		task := StartTaskWithTerminal("task message", term)
		task.AddLine("prefix", "line 1")
		task.AddLine("prefix", "line 2")
		task.AddLine("prefix", "line 3")
		task.Done()

		actual := buf.String()

		expected := strings.Join([]string{
			hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset,
			AnsiReset + AnsiFaint + "prefix" + AnsiReset + " line 1",
			AnsiReset + AnsiFaint + "prefix" + AnsiReset + " line 2",
			AnsiReset + AnsiFaint + "prefix" + AnsiReset + " line 3",
			checkmarkIcon + " task message" + AnsiFaint + "......" + AnsiReset + AnsiGreen + "OK" + AnsiReset + "\n",
		}, "\n")
		assert.Equal(t, expected, actual)
	})

	t.Run("UpdateLine", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		term := &stubTerminal{buf: buf, width: 25}

		t.Run("linebreak if no line is added", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line")
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("updates lines with carriage returns", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line version 1")
			task.UpdateLine("prefix", "updating line version 2")
			task.UpdateLine("prefix", "updating line version 3")
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line version 1",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line version 2",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line version 3",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("linebreak after preceding added line", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.AddLine("prefix", "preceding added line")
			task.UpdateLine("prefix", "updating line")
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				AnsiReset + AnsiFaint + "prefix" + AnsiReset + " preceding added line" + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("linebreak before following added line", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line")
			task.AddLine("prefix", "following added line")
			task.Done()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line" + "\n",
				AnsiReset + AnsiFaint + "prefix" + AnsiReset + " following added line" + "\n",
				checkmarkIcon + " task message" + AnsiFaint + "......" + AnsiReset + AnsiGreen + "OK" + AnsiReset + "\n",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("linebreak before task done", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line")
			task.Done()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line" + "\n",
				checkmarkIcon + " task message" + AnsiFaint + "......" + AnsiReset + AnsiGreen + "OK" + AnsiReset + "\n",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("linebreak before task error", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line")
			task.Error()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line" + "\n",
				errorcross + " task message" + AnsiFaint + "......" + AnsiReset + AnsiRed + "Error" + AnsiReset + "\n",
			}, "")
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("EndUpdateLine", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		term := &stubTerminal{buf: buf, width: 25}

		t.Run("adds a single linebreak after updating line", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line that is ended")
			task.EndUpdateLine()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line that is ended\n",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("is idempontent", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.UpdateLine("prefix", "updating line that is ended")
			task.EndUpdateLine()
			task.EndUpdateLine()
			task.EndUpdateLine()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				"\r" + AnsiReset + AnsiFaint + "prefix" + AnsiReset + " updating line that is ended\n",
			}, "")
			assert.Equal(t, expected, actual)
		})

		t.Run("does not print extra newline after adding line", func(t *testing.T) {
			buf.Reset()

			task := StartTaskWithTerminal("task message", term)
			task.AddLine("prefix", "added line")
			task.EndUpdateLine()
			actual := buf.String()

			expected := strings.Join([]string{
				hourglassIcon + " task message" + AnsiFaint + "......" + AnsiReset + "\n",
				AnsiReset + AnsiFaint + "prefix" + AnsiReset + " added line\n",
			}, "")
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Progress", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		term := &stubTerminal{buf: buf, width: 19} // width: task message + 10 spaces for progress + max suffix length

		task := StartTaskWithTerminal("message", term)
		for i := float32(0.0); i <= 100.0; i += 20.0 {
			task.Progress(i)
		}
		actual := buf.String()

		expected := strings.Join([]string{
			hourglassIcon + " message" + AnsiFaint + "....." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "....." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "#...." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "##..." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "###.." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "####." + AnsiReset,
			"\r" + hourglassIcon + " message" + AnsiFaint + "#####" + AnsiReset,
		}, "")
		assert.Equal(t, expected, actual)
	})

	t.Run("StartTask", func(t *testing.T) {
		t.Run("writes expected output to stdout", func(t *testing.T) {
			stdoutR := test.RunWithPatchedStdout(t, func() {
				StartTask("message")
			})

			actual, err := io.ReadAll(stdoutR)
			assert.NoError(t, err)
			expected := hourglassIcon + " message" + AnsiFaint + ".............................................." + AnsiReset
			assert.Equal(t, expected, string(actual))
		})
	})
}
