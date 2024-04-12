package wizard

import (
	"strings"

	"numerous/cli/tool"
)

type surveyAnswers struct {
	Name             string
	Description      string
	LibraryName      string
	AppFile          string
	RequirementsFile string
}

func (s surveyAnswers) appendAnswersToApp(a *tool.Tool) {
	a.Name = s.Name
	a.Description = s.Description
	a.Library, _ = tool.GetLibraryByName(s.LibraryName)
	a.AppFile = strings.Trim(s.AppFile, " 	")
	a.RequirementsFile = strings.Trim(s.RequirementsFile, " 	")
}

func fromApp(a *tool.Tool) *surveyAnswers {
	return &surveyAnswers{
		Name:             a.Name,
		Description:      a.Description,
		LibraryName:      a.Library.Name,
		AppFile:          a.AppFile,
		RequirementsFile: a.RequirementsFile,
	}
}
