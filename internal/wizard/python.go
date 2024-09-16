package wizard

import (
	"errors"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

type pythonAnswers struct {
	Library          manifest.Library
	AppFile          string
	RequirementsFile string
}

func (p pythonAnswers) ToManifest() (*manifest.Python, error) {
	return &manifest.Python{
		Library:          p.Library,
		AppFile:          strings.Trim(p.AppFile, " 	"),
		RequirementsFile: strings.Trim(p.RequirementsFile, " 	"),
		Port:             p.Library.Port,
	}, nil
}

func pythonWizard(asker Asker, libraryName string, p *manifest.Python) (pythonAnswers, error) {
	lib, err := manifest.GetLibraryByName(libraryName)
	if err != nil {
		return pythonAnswers{}, err
	}
	ps := pythonAnswers{Library: lib}
	qs := []*survey.Question{}

	if p == nil || p.AppFile == "" {
		qs = append(qs, getFileQuestion("AppFile", "Provide the path to your app:", "app.py", ".py"))
	} else {
		ps.AppFile = p.AppFile
	}

	if p == nil || p.RequirementsFile == "" {
		q := getFileQuestion("RequirementsFile", "Provide the path to your requirements file:", "requirements.txt", ".txt")
		qs = append(qs, q)
	} else {
		ps.RequirementsFile = p.RequirementsFile
	}

	if err := asker.Ask(qs, &ps); errors.Is(err, terminal.InterruptErr) {
		return ps, ErrStopInit
	}

	return ps, nil
}
