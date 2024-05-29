package wizard

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// allows dependency injection of the standard input for testing the survey
func UseOrCreateAppFolder(folderPath string, in terminal.FileReader) (bool, error) {
	absPath, err := absPath(folderPath)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return createFolderSurvey(absPath, in)
	}

	return confirmFolderSurvey(absPath, in)
}

func createFolderSurvey(folderPath string, in terminal.FileReader) (bool, error) {
	var confirm bool

	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Create new folder '%s'? (default: yes)", folderPath),
		Default: true,
	}

	err := survey.AskOne(prompt, &confirm, func(options *survey.AskOptions) error { options.Stdio.In = in; return nil })
	if err != nil {
		return false, err
	}

	if confirm {
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			return false, err
		}
	}

	return confirm, nil
}

func confirmFolderSurvey(folderPath string, in terminal.FileReader) (bool, error) {
	var confirm bool

	msg := fmt.Sprintf("Use the existing folder %s for your app? (default: yes)", folderPath)
	prompt := &survey.Confirm{Message: msg, Default: true}
	err := survey.AskOne(prompt, &confirm, func(options *survey.AskOptions) error { options.Stdio.In = in; return nil })
	if err != nil {
		return false, err
	}

	return confirm, nil
}

func absPath(p string) (string, error) {
	if path.IsAbs(p) {
		return p, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Clean(path.Join(wd, p)), nil
}
