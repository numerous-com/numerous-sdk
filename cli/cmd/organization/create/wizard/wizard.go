package wizard

import (
	"numerous/cli/auth"

	"github.com/AlecAivazis/survey/v2"
)

func RunOrganizationCreateWizard(name *string, user auth.User) error {
	questions := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Allowed inputs: a-z, A-Z, 0-9, \"-\", \" \"\nName of the organization:",
				Default: "user's organization",
			},
			Validate: survey.Required,
		},
	}

	answers := struct{ Name string }{}
	if err := survey.Ask(questions, &answers); err != nil {
		return err
	}

	*name = answers.Name

	return nil
}
