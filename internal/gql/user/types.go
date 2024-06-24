package user

import "numerous.com/cli/internal/gql/organization"

type User struct {
	FullName    string
	Memberships []organization.OrganizationMembership
}
