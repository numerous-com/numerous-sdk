package wizard

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"numerous.com/cli/internal/manifest"
)

type AppAnswers struct {
	Name        string
	Description string
	LibraryKey  string
	LibraryName string
}

func (s AppAnswers) ToManifestApp() manifest.App {
	return manifest.NewApp(s.Name, s.Description)
}

func appWizard(asker Asker, as AppAnswers) (AppAnswers, error) {
	qs := []*survey.Question{}

	if as.Name == "" {
		q := textQuestion("Name", "Name your app:", true)
		qs = append(qs, q)
	}

	if as.Description == "" {
		q := textQuestion("Description", "Provide a short description for your app:", false)
		qs = append(qs, q)
	}

	if as.LibraryKey == "" {
		q := libraryQuestion("LibraryName", "Select which app library you are using:")
		qs = append(qs, q)
	}

	if err := asker.Ask(qs, &as); errors.Is(err, terminal.InterruptErr) {
		return as, ErrStopInit
	}

	return as, nil
}

var dockerfileLibraryName = "Dockerfile"

func libraryQuestion(name, prompt string) *survey.Question {
	libraryNames := []string{}
	for _, lib := range manifest.SupportedLibraries {
		libraryNames = append(libraryNames, lib.Name)
	}

	libraryNames = append(libraryNames, dockerfileLibraryName)

	return &survey.Question{
		Name: name,
		Prompt: &survey.Select{
			Message: prompt,
			Options: libraryNames,
		},
	}
}

func textQuestion(name, prompt string, required bool) *survey.Question {
	q := &survey.Question{
		Name:   name,
		Prompt: &survey.Input{Message: prompt},
	}
	if required {
		q.Validate = survey.Required
	}

	return q
}
