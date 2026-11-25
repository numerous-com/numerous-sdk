package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrganizationURL(t *testing.T) {
	slug := "organization-slug"
	url := GetOrganizationURL(slug)

	assert.Equal(t, "https://www.numerous.com/app/organization/"+slug, url)
}

func TestGetAppURL(t *testing.T) {
	orgSlug := "organization-slug"
	appSlug := "app-slug"
	url := GetAppURL(orgSlug, appSlug)

	assert.Equal(t, "https://www.numerous.com/app/organization/"+orgSlug+"/private/"+appSlug, url)
}

func TestGetAppMCPURL(t *testing.T) {
	assert.Equal(t, "https://my-app.api.numerous.com/mcp", GetAppMCPURL("my-app", ""))
	assert.Equal(t, "https://my-app.api.numerous.com/custom/mcp", GetAppMCPURL("my-app", "custom/mcp/"))
	assert.Equal(t, "https://my-app.api.numerous.com/", GetAppMCPURL("my-app", "/"))
}
