package wizard

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

type DockerAnswers struct {
	Dockerfile string
	Context    string
}

func (d DockerAnswers) ToManifest() *manifest.Docker {
	if d.Dockerfile == "" {
		return nil
	}

	return &manifest.Docker{
		Dockerfile: d.Dockerfile,
		Context:    d.Context,
	}
}

func dockerWizard(asker Asker, pas DockerAnswers) (DockerAnswers, error) {
	qs := []*survey.Question{}

	if pas.Dockerfile == "" {
		q := getFileQuestion("Dockerfile", "Select the Dockerfile", "Dockerfile", "")
		qs = append(qs, q)
	}

	if pas.Context == "" {
		q := getFolderQuestion("Context", "Select the Docker build context folder", ".")
		qs = append(qs, q)
	}

	if err := asker.Ask(qs, &pas); errors.Is(err, terminal.InterruptErr) {
		return pas, ErrStopInit
	}

	return pas, nil
}
