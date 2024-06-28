package links

const baseURL = "https://www.numerous.com/app"

func GetOrganizationURL(orgSlug string) string {
	return baseURL + "/organization/" + orgSlug
}

func GetAppURL(orgSlug, appSlug string) string {
	return GetOrganizationURL(orgSlug) + "/private/" + appSlug
}
