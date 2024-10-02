package requirements

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

var testCases = []string{
	"utf-8-bom-crlf",
	"utf-8-bom-lf",
	"utf-8-no-bom-crlf",
	"utf-8-no-bom-lf",
	"utf-16-le-crlf",
	"utf-16-le-lf",
	"utf-16-be-crlf",
	"utf-16-be-lf",
	"utf-32-le-crlf",
	"utf-32-le-lf",
	"utf-32-be-crlf",
	"utf-32-be-lf",
	"last-line-empty-utf-16-be-crlf",
}

func TestRead(t *testing.T) {
	expectedLinesContent, err := os.ReadFile("testdata/lines.txt")
	require.NoError(t, err)
	expectedLines := strings.Split(string(expectedLinesContent), "\n")

	t.Run("reads expected lines", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc, func(t *testing.T) {
				f, err := os.Open("testdata/" + tc + ".txt")
				require.NoError(t, err)

				req, err := Read(f)
				f.Close()

				if strings.Contains(tc, "last-line-empty") {
					expectedLines = append(expectedLines, "")
				}

				assert.NoError(t, err)
				if assert.NotNil(t, req) {
					expectedCrlf := strings.Contains(tc, "crlf")
					assert.Equal(t, expectedCrlf, req.crlf)
					assert.Equal(t, expectedLines, req.lines)
				}
			})
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("writes expected requirements file", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc, func(t *testing.T) {
				filename := "testdata/" + tc + ".txt"
				expected, err := os.ReadFile(filename)
				require.NoError(t, err)

				f, err := os.Open(filename)
				require.NoError(t, err)
				req, err := Read(f)
				require.NoError(t, err)

				f.Close()
				actualPath := t.TempDir() + "/requirements.txt"
				actualFile, err := os.Create(actualPath)
				require.NoError(t, err)

				err = req.Write(actualFile)

				assert.NoError(t, err)
				if assert.NotNil(t, req) {
					test.AssertFileContent(t, actualPath, expected)
				}
			})
		}
	})
}

func TestAdd(t *testing.T) {
	for _, tc := range []struct {
		name     string
		initial  string
		toAdd    []string
		expected string
	}{
		{
			name:     "adds to empty requirements",
			initial:  "",
			toAdd:    []string{"dep1", "dep2"},
			expected: "dep1\ndep2\n",
		},
		{
			name:     "does not add newline when nothing added",
			initial:  "dep1\ndep2",
			toAdd:    []string{"dep1", "dep2"},
			expected: "dep1\ndep2",
		},
		{
			name:     "does not add existing dependency again",
			initial:  "dep2",
			toAdd:    []string{"dep1", "dep2"},
			expected: "dep2\ndep1\n",
		},
		{
			name:     "does not add existing dependency again with version",
			initial:  "dep2==v1.2.3",
			toAdd:    []string{"dep1", "dep2"},
			expected: "dep2==v1.2.3\ndep1\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rbuf := bytes.NewBufferString(tc.initial)
			req, err := Read(rbuf)
			require.NoError(t, err)

			for _, r := range tc.toAdd {
				req.Add(r)
			}

			wbuf := bytes.NewBuffer(nil)
			err = req.Write(wbuf)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, wbuf.String())
		})
	}
}
