package config

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

func TestPrint(t *testing.T) {
	c := Config{OrganizationSlug: "test-organization-slug"}

	stdoutR := test.RunWithPatchedStdout(t, func() {
		c.Print()
	})

	data, err := io.ReadAll(stdoutR)
	if assert.NoError(t, err) {
		expected := "organization = \"test-organization-slug\"\n"
		assert.Equal(t, expected, string(data))
	}
}
