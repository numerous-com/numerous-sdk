package config

func OrganizationSlug() string {
	c := Config{}

	if err := c.Load(); err != nil {
		return ""
	}

	return c.OrganizationSlug
}
