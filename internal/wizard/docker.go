package wizard

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

type dockerAnswers struct {
	Dockerfile string
	Context    string
}

func (p dockerAnswers) ToManifest() *manifest.Docker {
	return &manifest.Docker{
		Dockerfile: p.Dockerfile,
		Context:    p.Context,
	}
}

func dockerWizard(asker Asker, p *manifest.Docker) (dockerAnswers, error) {
	var pas dockerAnswers
	qs := []*survey.Question{}

	if p == nil || p.Dockerfile == "" {
		q := getFileQuestion("Dockerfile", "Select the Dockerfile", "Dockerfile", "")
		qs = append(qs, q)
	} else {
		pas.Dockerfile = p.Dockerfile
	}

	if p == nil || p.Context == "" {
		q := getFolderQuestion("Context", "Select the Docker build context folder", ".")
		qs = append(qs, q)
	} else {
		pas.Context = p.Context
	}

	if err := asker.Ask(qs, &pas); errors.Is(err, terminal.InterruptErr) {
		return pas, ErrStopInit
	}

	return pas, nil
}
