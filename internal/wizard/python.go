package wizard

import (
	"errors"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

type PythonAnswers struct {
	Library          manifest.Library
	AppFile          string
	RequirementsFile string
}

func (p PythonAnswers) ToManifest() *manifest.Python {
	if p.Library.Key == "" {
		return nil
	}

	return &manifest.Python{
		Library:          p.Library,
		AppFile:          strings.Trim(p.AppFile, " 	"),
		RequirementsFile: strings.Trim(p.RequirementsFile, " 	"),
		Port:             p.Library.Port,
	}
}

func pythonWizard(asker Asker, libraryName string, ps PythonAnswers) (PythonAnswers, error) {
	lib, err := manifest.GetLibraryByName(libraryName)
	if err != nil {
		return PythonAnswers{}, err
	}

	ps.Library = lib
	qs := []*survey.Question{}

	if ps.AppFile == "" {
		qs = append(qs, getFileQuestion("AppFile", "Provide the path to your app:", "app.py", ".py"))
	}

	if ps.RequirementsFile == "" {
		q := getFileQuestion("RequirementsFile", "Provide the path to your requirements file:", "requirements.txt", ".txt")
		qs = append(qs, q)
	}

	if err := asker.Ask(qs, &ps); errors.Is(err, terminal.InterruptErr) {
		return ps, ErrStopInit
	}

	return ps, nil
}
