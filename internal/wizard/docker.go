package wizard

import (
	"errors"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

var (
	ErrPortInvalidValue = errors.New("invalid value")
	ErrPortMustBeNumber = errors.New("must be a number")
)

type DockerAnswers struct {
	Dockerfile string
	Context    string
	Port       int
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

	if pas.Port == 0 {
		q := portQuestion()
		qs = append(qs, q)
	}

	if err := asker.Ask(qs, &pas); errors.Is(err, terminal.InterruptErr) {
		return pas, ErrStopInit
	}

	return pas, nil
}

func portQuestion() *survey.Question {
	q := &survey.Question{
		Name:   "Port",
		Prompt: &survey.Input{Message: "Select the port number of your app", Default: "8000"},
	}

	q.Validate = func(ans interface{}) error {
		s, ok := ans.(string)
		if !ok {
			return ErrPortInvalidValue
		}

		_, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return ErrPortMustBeNumber
		}

		return nil
	}

	return q
}
