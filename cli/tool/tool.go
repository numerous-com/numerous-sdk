package tool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const ToolIDFileName string = ".tool_id.txt"

var ErrToolIDNotFound = fmt.Errorf("%s not found", ToolIDFileName)

type Tool struct {
	Name             string
	Description      string
	Library          Library
	Python           string
	AppFile          string
	RequirementsFile string
	CoverImage       string
}

func ToolIDExistsInCurrentDir(t *Tool, basePath string) (bool, error) {
	_, err := os.Stat(filepath.Join(basePath, ToolIDFileName))
	exists := !errors.Is(err, os.ErrNotExist)

	return exists, err
}

func DeleteTool(basePath string) error {
	return os.Remove(filepath.Join(basePath, ToolIDFileName))
}

func ReadToolID(basePath string) (string, error) {
	toolID, err := os.ReadFile(filepath.Join(basePath, ToolIDFileName))
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrToolIDNotFound
	} else if err != nil {
		return "", err
	}

	return string(toolID), nil
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
