package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertFileContent(t *testing.T, filename string, expected []byte) bool {
	t.Helper()

	if !assert.FileExists(t, filename) {
		return false
	}

	actual, err := os.ReadFile(filename)
	if !assert.NoError(t, err) {
		return false
	}

	return assert.Equal(t, expected, actual)
}
