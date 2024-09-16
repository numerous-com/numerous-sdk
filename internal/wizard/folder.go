package wizard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

var (
	ErrMustBeDirectory = errors.New("file must be a directory")
	ErrNotString       = errors.New("not a string")
)

func UseOrCreateAppFolder(asker Asker, folderPath string) (bool, error) {
	absPath, err := absPath(folderPath)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return createFolderSurvey(asker, absPath)
	}

	return confirmFolderSurvey(asker, absPath)
}

func createFolderSurvey(asker Asker, folderPath string) (bool, error) {
	var confirm bool

	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Create new folder '%s'? (default: yes)", folderPath),
		Default: true,
	}

	err := asker.AskOne(prompt, &confirm)
	if errors.Is(err, terminal.InterruptErr) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if confirm {
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			return false, err
		}
	}

	return confirm, nil
}

func confirmFolderSurvey(asker Asker, folderPath string) (bool, error) {
	var confirm bool

	msg := fmt.Sprintf("Use the existing folder %s for your app? (default: yes)", folderPath)
	prompt := &survey.Confirm{Message: msg, Default: true}
	err := asker.AskOne(prompt, &confirm)
	if errors.Is(err, terminal.InterruptErr) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return confirm, nil
}

func absPath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(wd, p), nil
}

func getFolderQuestion(name, prompt, defaultPath string) *survey.Question {
	return &survey.Question{
		Name: name,
		Prompt: &survey.Input{
			Message: prompt,
			Default: defaultPath,
			Suggest: suggestFolder,
		},
		Validate: func(ans interface{}) error {
			folder, ok := ans.(string)
			if !ok {
				return ErrNotString
			}

			f, err := os.Stat(folder)
			if err != nil {
				return err
			}

			if !f.IsDir() {
				return ErrMustBeDirectory
			}

			return nil
		},
		Transform: cleanPath,
	}
}

func suggestFolder(toComplete string) []string {
	matches, _ := filepath.Glob(toComplete + "*")
	var paths []string
	for _, match := range matches {
		f, _ := os.Stat(match)
		if f.IsDir() {
			paths = append(paths, match+string(os.PathSeparator))
		}
	}

	return paths
}
