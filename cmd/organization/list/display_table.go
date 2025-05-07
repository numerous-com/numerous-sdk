package list

import (
	"fmt"

	"numerous.com/cli/internal/gql/organization"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			Foreground(lipgloss.Color("2")).
			PaddingLeft(1).
			PaddingRight(1)
	rowStyle = lipgloss.NewStyle().Padding(0, 1)
)

func setupTable(organizations []organization.OrganizationMembership, configuredOrganizationSlug string) *table.Table {
	columns := []string{"Name", "Slug", "Role", "ID", "Default"}
	var rows [][]string
	for _, o := range organizations {
		isDefault := ""
		if o.Organization.Slug == configuredOrganizationSlug {
			isDefault = "*"
		}
		rows = append(rows, []string{
			o.Organization.Name,
			o.Organization.Slug,
			string(o.Role),
			o.Organization.ID,
			isDefault,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		BorderRow(true).
		Headers(columns...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			if row == 0 {
				style = headerStyle
			} else {
				style = rowStyle
			}

			return style
		})

	return t
}

func displayTable(organizations []organization.OrganizationMembership, configuredOrganizationSlug string) {
	fmt.Println(setupTable(organizations, configuredOrganizationSlug))
}
