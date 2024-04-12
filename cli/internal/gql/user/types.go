package user

import "numerous/cli/internal/gql/organization"

type User struct {
	FullName    string
	Memberships []organization.OrganizationMembership
}
