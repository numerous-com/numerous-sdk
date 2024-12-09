package list

import (
	"fmt"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/gql/organization"
)

func displayList(memberships []organization.OrganizationMembership, configuredOrganizationSlug string) {
	first := true
	for _, membership := range memberships {
		if first {
			first = false
		} else {
			fmt.Println()
		}

		displayMembership(membership, configuredOrganizationSlug)
	}
}

func displayMembership(membership organization.OrganizationMembership, configuredOrganizationSlug string) {
	o := membership.Organization
	role := membership.Role

	if configuredOrganizationSlug == o.Slug {
		fmt.Printf("Name:    %s\n", o.Name)
		fmt.Printf("Slug:    %s\n", o.Slug)
		fmt.Printf("Role:    %s\n", role)
		fmt.Printf("ID:      %s\n", o.ID)
		fmt.Println("Default: " + output.AnsiGreen + "Yes" + output.AnsiReset)
	} else {
		fmt.Printf("Name: %s\n", o.Name)
		fmt.Printf("Slug: %s\n", o.Slug)
		fmt.Printf("Role: %s\n", role)
		fmt.Printf("ID:   %s\n", o.ID)
	}
}
