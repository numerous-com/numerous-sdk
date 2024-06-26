package initialize

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"numerous.com/cli/assets"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/manifest"
)

const EnvFileName string = ".env"

func BootstrapFiles(m *manifest.Manifest, toolID string, basePath string) error {
	manifestToml, err := m.ToToml()
	if err != nil {
		output.PrintErrorDetails("Error encoding manifest file", err)

		return err
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
	if err = addToGitIgnore(basePath, gitignoreLines); err != nil {
		return err
	}

	appFilePath := filepath.Join(basePath, m.AppFile)
	if err = createAndWriteIfFileNotExist(appFilePath, m.Library.DefaultAppFile()); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(basePath, m.RequirementsFile)
	if err = createFile(requirementsFilePath); err != nil {
		return err
	}

	if err := bootstrapRequirements(m, basePath); err != nil {
		return err
	}
	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, m.CoverImage)); err != nil {
		return err
	}

	if err = createAndWriteIfFileNotExist(filepath.Join(basePath, manifest.ManifestPath), manifestToml); err != nil {
		return err
	}

	return nil
}

func bootstrapRequirements(m *manifest.Manifest, basePath string) error {
	requirementsPath := filepath.Join(basePath, m.RequirementsFile)
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

	for _, requirement := range m.Library.Requirements {
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
