package wizard

import (
	"fmt"
	"os"

	"numerous/cli/manifest"

	"github.com/AlecAivazis/survey/v2"
)

type InitWizardOptions struct {
	Name             string
	Description      string
	LibraryKey       string
	AppFile          string
	RequirementsFile string
}

func RunInitAppWizard(projectFolderPath string, m *manifest.Manifest) (bool, error) {
	questions := getQuestions(m)
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

	answers := answersFromManifest(m)
	if err := survey.Ask(questions, &answers); err != nil {
		return false, err
	}

	answers.updateManifest(m)

	return true, nil
}

func getQuestions(m *manifest.Manifest) []*survey.Question {
	qs := []*survey.Question{}

	if m.Name == "" {
		q := getTextQuestion("Name", "Name your app:", true)
		qs = append(qs, q)
	}

	if m.Description == "" {
		q := getTextQuestion("Description", "Provide a short description for your app:", false)
		qs = append(qs, q)
	}

	if m.Library.Key == "" {
		q := getLibraryQuestion("LibraryName", "Select which app library you are using:")
		qs = append(qs, q)
	}

	if m.AppFile == "" {
		qs = append(qs, getFileQuestion("AppFile",
			"Provide the path to your app:", "app.py", ".py"))
	}

	if m.RequirementsFile == "" {
		q := getFileQuestion("RequirementsFile", "Provide the path to your requirements file:", "requirements.txt", ".txt")
		qs = append(qs, q)
	}

	return qs
}
