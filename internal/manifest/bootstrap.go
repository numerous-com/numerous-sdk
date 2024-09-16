package manifest

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"numerous.com/cli/assets"
	"numerous.com/cli/internal/dir"
)

const EnvFileName string = ".env"

var ErrEncodingManifest = errors.New("error encoding manifest")

func (m *Manifest) BootstrapFiles(toolID string, basePath string) error {
	manifestToml, err := m.ToTOML()
	if err != nil {
		return ErrEncodingManifest
	}

	if toolID != "" {
		if err = createAppIDFile(basePath, toolID); err != nil {
			return err
		}
	}

	gitignoreLines := []string{"# added by numerous init\n", EnvFileName}
	if toolID != "" {
		gitignoreLines = append(gitignoreLines, dir.AppIDFileName)
	}

	if err := addToGitIgnore(basePath, gitignoreLines); err != nil {
		return err
	}

	if m.Python == nil {
		return ErrNoPythonAppConfig
	}

	appFilePath := filepath.Join(basePath, m.Python.AppFile)
	if err = createAndWriteIfFileNotExist(appFilePath, m.Python.Library.DefaultAppFile()); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(basePath, m.Python.RequirementsFile)
	if err = createFile(requirementsFilePath); err != nil {
		return err
	}

	if err := bootstrapRequirements(m, basePath); err != nil {
		return err
	}
	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, m.CoverImage)); err != nil {
		return err
	}

	if err = createAndWriteIfFileNotExist(filepath.Join(basePath, ManifestPath), manifestToml); err != nil {
		return err
	}

	return nil
}

func bootstrapRequirements(m *Manifest, basePath string) error {
	requirementsPath := filepath.Join(basePath, m.Python.RequirementsFile)
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

	for _, requirement := range m.Python.Library.Requirements {
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
