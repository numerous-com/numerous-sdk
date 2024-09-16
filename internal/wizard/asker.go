package wizard

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
)

var ErrNoStubAnswer = errors.New("no stub answer")

type Asker interface {
	Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error
	AskOne(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error
}

var _ Asker = &SurveyAsker{}

type SurveyAsker struct{}

// AskOne implements Asker.
func (a *SurveyAsker) AskOne(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return survey.AskOne(prompt, response, opts...)
}

// Ask implements Asker.
func (a *SurveyAsker) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	return survey.Ask(qs, response, opts...)
}

var _ Asker = &StubAsker{}

// Used for stubbing out the Asker interface. Use as a map from survey.Confirm
// messages or survey.Question names to values. Errors in the map are returned
// as errors from the input process.
type StubAsker map[string]interface{}

// AskOne implements Asker.
func (a StubAsker) AskOne(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if conf, ok := prompt.(*survey.Confirm); ok {
		answer, ok := a[conf.Message]
		if !ok {
			return ErrNoStubAnswer
		}

		if err, ok := answer.(error); ok {
			return err
		}

		return core.WriteAnswer(response, "", answer)
	}

	return ErrNoStubAnswer
}

// Ask implements Asker.
func (a StubAsker) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	for _, q := range qs {
		answer, ok := a[q.Name]
		if !ok {
			return ErrNoStubAnswer
		}

		if err, ok := answer.(error); ok {
			return err
		}

		if err := core.WriteAnswer(response, q.Name, answer); err != nil {
			return err
		}
	}

	return nil
}
