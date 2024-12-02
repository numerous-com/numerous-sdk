package list

import (
	"fmt"

	"numerous.com/cli/internal/gql/organization"
)

func displayList(memberships []organization.OrganizationMembership) {
	first := true
	for _, membership := range memberships {
		if first {
			first = false
		} else {
			fmt.Println()
		}

		displayMembership(membership)
	}
}

func displayMembership(membership organization.OrganizationMembership) {
	o := membership.Organization
	role := membership.Role

	fmt.Printf("Name: %s\n", o.Name)
	fmt.Printf("Slug: %s\n", o.Slug)
	fmt.Printf("Role: %s\n", role)
	fmt.Printf("ID:   %s\n", o.ID)
}
