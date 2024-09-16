package manifest

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
)

func (p Python) bootstrapFiles(basePath string) error {
	appFilePath := filepath.Join(basePath, p.AppFile)
	if err := createAndWriteIfFileNotExist(appFilePath, p.Library.DefaultAppFile()); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(basePath, p.RequirementsFile)
	if err := createFile(requirementsFilePath); err != nil {
		return err
	}

	if err := p.bootstrapRequirements(basePath); err != nil {
		return err
	}

	return nil
}

func (p Python) bootstrapRequirements(basePath string) error {
	requirementsPath := filepath.Join(basePath, p.RequirementsFile)
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		return err
	}
	var filePermission fs.FileMode = 0o600
	file, err := os.OpenFile(requirementsPath, os.O_APPEND|os.O_WRONLY, filePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, requirement := range p.Library.Requirements {
		if err := addRequirementToFile(file, content, requirement); err != nil {
			return err
		}
	}

	return nil
}

func addRequirementToFile(f *os.File, content []byte, requirement string) error {
	if bytes.Contains(content, []byte(requirement)) {
		return nil
	}

	// If it comes after content without newline, add newline
	if len(content) != 0 && !bytes.HasSuffix(content, []byte("\n")) {
		requirement = "\n" + requirement
	}

	if _, err := f.WriteString(requirement + "\n"); err != nil {
		return err
	}

	return nil
}
