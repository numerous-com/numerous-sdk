package wizard

import (
	"errors"
	"fmt"

	"numerous.com/cli/internal/manifest"
)

var ErrStopInit = errors.New("stop app init")

func questionsNeeded(m *manifest.Manifest) bool {
	if m.Name == "" {
		return true
	}

	if m.Docker == nil && m.Python != nil {
		if m.Python.AppFile == "" {
			return true
		}

		if m.Python.Library.Key == "" {
			return true
		}
	}

	if m.Docker == nil {
		return true
	}

	return false
}

func Run(asker Asker, projectFolderPath string, m *manifest.Manifest) error {
	if !questionsNeeded(m) {
		return ErrStopInit
	}

	fmt.Println("Hi there, welcome to Numerous.")
	fmt.Println("We're happy you're here!")
	fmt.Println("Let's get started by entering basic information about your app.")

	continueWizard, err := UseOrCreateAppFolder(asker, projectFolderPath)
	if err != nil {
		return err
	} else if !continueWizard {
		return ErrStopInit
	}

	appAnswers, err := appWizard(asker, m)
	if err != nil {
		return err
	}
	appAnswers.UpdateManifest(m)

	if appAnswers.LibraryName == dockerfileLibraryName {
		dockerAnswers, err := dockerWizard(asker, m.Docker)
		if err != nil {
			return err
		}
		m.Docker = dockerAnswers.ToManifest()
	} else {
		pythonAnswers, err := pythonWizard(asker, appAnswers.LibraryName, m.Python)
		if err != nil {
			return err
		}

		pythonManifest, err := pythonAnswers.ToManifest()
		if err != nil {
			return err
		}
		m.Python = pythonManifest
	}

	return nil
}
