package wizard

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
)

func getFileQuestion(name, prompt, defaultPath, fileExtension string) *survey.Question {
	return &survey.Question{
		Name: name,
		Prompt: &survey.Input{
			Message: prompt,
			Default: defaultPath,
			Suggest: func(toComplete string) []string {
				return suggestPath(toComplete, fileExtension)
			},
		},
		Validate: func(ans interface{}) error {
			return validatePath(ans, fileExtension)
		},
		Transform: cleanPath,
	}
}

func suggestPath(toComplete string, fileExtension string) []string {
	matches, _ := filepath.Glob(toComplete + "*")
	var paths []string
	for _, match := range matches {
		f, _ := os.Stat(match)
		if f.IsDir() {
			paths = append(paths, match+string(os.PathSeparator))
		} else if filepath.Ext(match) == fileExtension {
			paths = append(paths, match)
		}
	}

	return paths
}

func validatePath(ans interface{}, fileExtension string) error {
	if err := survey.Required(ans); err != nil {
		return err
	}
	if filepath.Ext(fmt.Sprintf("%v", ans)) != fileExtension {
		return fmt.Errorf("input must be a %s file", fileExtension)
	}

	return nil
}

func cleanPath(path interface{}) interface{} {
	return filepath.Clean(fmt.Sprintf("%v", path))
}
