package initialize

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"numerous/cli/assets"
	"numerous/cli/cmd/output"
	"numerous/cli/manifest"
	"numerous/cli/tool"
)

const EnvFileName string = ".env"

func bootstrapFiles(t tool.Tool, toolID string, basePath string) error {
	manifestToml, err := manifest.FromTool(t).ToToml()
	if err != nil {
		output.PrintErrorDetails("Error encoding manifest file", err)

		return err
	}

	if err = createAppIDFile(basePath, toolID); err != nil {
		return err
	}

	if err = addToGitIgnore(basePath, []string{"# added by numerous init\n", tool.AppIDFileName, EnvFileName}); err != nil {
		return err
	}

	appFilePath := filepath.Join(basePath, t.AppFile)
	if err = createAndWriteIfFileNotExist(appFilePath, t.Library.DefaultAppFile()); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(basePath, t.RequirementsFile)
	if err = createFile(requirementsFilePath); err != nil {
		return err
	}

	if err := bootstrapRequirements(t, basePath); err != nil {
		return err
	}
	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, t.CoverImage)); err != nil {
		return err
	}

	if err = createAndWriteIfFileNotExist(filepath.Join(basePath, manifest.ManifestPath), manifestToml); err != nil {
		return err
	}

	return nil
}

func bootstrapRequirements(t tool.Tool, basePath string) error {
	requirementsPath := filepath.Join(basePath, t.RequirementsFile)
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

	for _, requirement := range t.Library.Requirements {
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
