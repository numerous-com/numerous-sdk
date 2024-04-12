package wizard

import (
	"fmt"
	"os"

	"numerous/cli/tool"

	"github.com/AlecAivazis/survey/v2"
)

func RunInitAppWizard(projectFolderPath string, a *tool.Tool) (bool, error) {
	questions := getQuestions(*a)
	if len(questions) == 1 && questions[0].Name == "Description" {
		return false, nil
	}

	fmt.Println("Hi there, welcome to Numerous.")
	fmt.Println("We're happy you're here!")
	fmt.Println("Let's get started by entering basic information about your app.")

	continueWizard, err := UseOrCreateAppFolder(projectFolderPath, os.Stdin)
	if err != nil {
		return false, err
	} else if !continueWizard {
		return false, nil
	}

	answers := fromApp(a)
	if err := survey.Ask(questions, answers); err != nil {
		return false, err
	}

	answers.appendAnswersToApp(a)

	return true, nil
}

func getQuestions(a tool.Tool) []*survey.Question {
	q := []*survey.Question{}

	if a.Name == "" {
		q = append(q, getTextQuestion("Name",
			"Name your app:",
			true))
	}
	if a.Description == "" {
		q = append(q, getTextQuestion("Description",
			"Provide a short description for your app:",
			false))
	}
	if a.Library.Key == "" {
		q = append(q, getLibraryQuestion("LibraryName",
			"Select which app library you are using:"))
	}
	if a.AppFile == "" {
		q = append(q, getFileQuestion("AppFile",
			"Provide the path to your app:", "app.py", ".py"))
	}
	if a.RequirementsFile == "" {
		q = append(q, getFileQuestion("RequirementsFile",
			"Provide the path to your requirements file:",
			"requirements.txt", ".txt"))
	}

	return q
}
