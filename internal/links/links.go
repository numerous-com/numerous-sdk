package links

import "strings"

const (
	baseURL           = "https://www.numerous.com/app"
	defaultAppMCPPath = "/mcp"
)

func GetOrganizationURL(orgSlug string) string {
	return baseURL + "/organization/" + orgSlug
}

func GetAppURL(orgSlug, appSlug string) string {
	return GetOrganizationURL(orgSlug) + "/private/" + appSlug
}

func GetAppMCPURL(appSlug, mcpPath string) string {
	cleaned := strings.TrimSpace(mcpPath)
	if cleaned == "" {
		cleaned = defaultAppMCPPath
	}
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	if cleaned != "/" {
		cleaned = strings.TrimRight(cleaned, "/")
	}

	return "https://" + appSlug + ".api.numerous.com" + cleaned
}
