package list

import (
	"numerous.com/cli/internal/gql/app"

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

func setupTable(apps []app.App) *table.Table {
	columns := []string{"ID", "Name", "Description", "Created At ", "Shareable URL", "Public App"}
	publicAppColumnIdx := 5
	var rows [][]string
	for _, a := range apps {
		rows = append(rows, []string{
			a.ID,
			a.Name,
			a.Description,
			a.CreatedAt.Local().Format("2006-Jan-02"),
			a.SharedURL,
			getPublicEmoji(a.PublicURL),
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
			if col == publicAppColumnIdx {
				style = style.AlignHorizontal(lipgloss.Center)
			}

			return style
		})

	return t
}

func getPublicEmoji(publicURL string) string {
	if publicURL != "" {
		return "✅"
	}

	return ""
}
