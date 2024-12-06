package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationSlug(t *testing.T) {
	configBaseDir = t.TempDir()
	testOrganizationSlug := "test-organization-slug"
	c := Config{OrganizationSlug: testOrganizationSlug}
	require.NoError(t, c.Save())

	assert.Equal(t, testOrganizationSlug, OrganizationSlug())
}
