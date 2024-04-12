package wizard

import "github.com/AlecAivazis/survey/v2"

func getTextQuestion(name, prompt string, required bool) *survey.Question {
	q := &survey.Question{
		Name:   name,
		Prompt: &survey.Input{Message: prompt},
	}
	if required {
		q.Validate = survey.Required
	}

	return q
}
