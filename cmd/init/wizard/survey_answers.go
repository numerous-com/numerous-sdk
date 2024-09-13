package wizard

import (
	"strings"

	"numerous.com/cli/internal/manifest"
)

type surveyAnswers struct {
	Name             string
	Description      string
	LibraryName      string
	AppFile          string
	RequirementsFile string
}

func (s surveyAnswers) updateManifest(m *manifest.Manifest) {
	lib, _ := manifest.GetLibraryByName(s.LibraryName) // TODO: handle error here?
	m.Name = s.Name
	m.Description = s.Description
	m.Python.Library = lib
	m.Python.AppFile = strings.Trim(s.AppFile, " 	")
	m.Python.RequirementsFile = strings.Trim(s.RequirementsFile, " 	")
	m.Python.Port = lib.Port
}

func answersFromManifest(m *manifest.Manifest) surveyAnswers {
	return surveyAnswers{
		Name:             m.Name,
		Description:      m.Description,
		LibraryName:      m.Python.Library.Name,
		AppFile:          m.Python.AppFile,
		RequirementsFile: m.Python.RequirementsFile,
	}
}
