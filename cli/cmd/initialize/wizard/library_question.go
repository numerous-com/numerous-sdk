package wizard

import (
	"numerous/cli/tool"

	"github.com/AlecAivazis/survey/v2"
)

func getLibraryQuestion(name, prompt string) *survey.Question {
	libraryNames := []string{}
	for _, lib := range tool.SupportedLibraries {
		libraryNames = append(libraryNames, lib.Name)
	}

	return &survey.Question{
		Name: name,
		Prompt: &survey.Select{
			Message: prompt,
			Options: libraryNames,
		},
	}
}
