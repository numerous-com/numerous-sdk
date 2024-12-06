package config

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

func TestPrint(t *testing.T) {
	stdoutR, stdoutW := test.PatchStdout(t)

	c := Config{OrganizationSlug: "test-organization-slug"}
	c.Print()

	assert.NoError(t, stdoutW.Close())
	data, err := io.ReadAll(stdoutR)

	if assert.NoError(t, err) {
		expected := "organization = \"test-organization-slug\"\n"
		assert.Equal(t, expected, string(data))
	}
}
