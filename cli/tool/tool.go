package tool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ToolIDFileName string = ".tool_id.txt" // supported for backwards compatibility
	AppIDFileName  string = ".app_id.txt"
)

var ErrAppIDNotFound = errors.New("app id not found")

type Tool struct {
	Name             string
	Description      string
	Library          Library
	Python           string
	AppFile          string
	RequirementsFile string
	CoverImage       string
}

func AppIDExistsInCurrentDir(t *Tool, basePath string) (bool, error) {
	_, err := os.Stat(filepath.Join(basePath, AppIDFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	_, err = os.Stat(filepath.Join(basePath, ToolIDFileName))
	if !errors.Is(err, os.ErrNotExist) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	// neither file exists
	return false, nil
}

func ReadAppID(basePath string) (string, error) {
	appID, err := os.ReadFile(filepath.Join(basePath, AppIDFileName))
	if err == nil {
		return string(appID), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	appID, err = os.ReadFile(filepath.Join(basePath, ToolIDFileName))
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrAppIDNotFound
	}

	if err != nil {
		return "", err
	}

	return string(appID), nil
}

func (t Tool) String() string {
	return fmt.Sprintf(`
Tool:
	name             %s
	description      %s
	library          %s
	python			 %s
	appFile          %s
	requirementsFile %s
	coverImage       %s
	`, t.Name, t.Description, t.Library.Key, t.Python, t.AppFile, t.RequirementsFile, t.CoverImage)
}
