package organization

type Organization struct {
	ID   string
	Name string
	Slug string
}

type Role string

const (
	Admin Role = "ADMIN"
	User  Role = "USER"
)

type OrganizationMembership struct {
	Role         Role
	Organization Organization
}
