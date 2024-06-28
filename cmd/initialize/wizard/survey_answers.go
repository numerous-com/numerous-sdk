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
	// TODO: handle error here?
	lib, _ := manifest.GetLibraryByName(s.LibraryName)
	m.Name = s.Name
	m.Description = s.Description
	m.Library = lib
	m.AppFile = strings.Trim(s.AppFile, " 	")
	m.RequirementsFile = strings.Trim(s.RequirementsFile, " 	")
	m.Port = lib.Port
}

func answersFromManifest(m *manifest.Manifest) surveyAnswers {
	return surveyAnswers{
		Name:             m.Name,
		Description:      m.Description,
		LibraryName:      m.Library.Name,
		AppFile:          m.AppFile,
		RequirementsFile: m.RequirementsFile,
	}
}
