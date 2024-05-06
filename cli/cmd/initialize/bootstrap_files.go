package initialize

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"numerous/cli/assets"
	"numerous/cli/manifest"
	"numerous/cli/tool"
)

func bootstrapFiles(t tool.Tool, toolID string, basePath string) (err error) {
	manifestToml, err := manifest.FromTool(t).ToToml()
	if err != nil {
		fmt.Println("Error encoding manifest file")

		return err
	}

	err = createAppIDFile(basePath, toolID)
	if err != nil {
		return err
	}
	if err = addToGitIgnore(basePath, "# added by numerous init\n"+tool.ToolIDFileName); err != nil {
		return err
	}

	if err = CreateAndWriteIfFileNotExist(filepath.Join(basePath, t.AppFile), t.Library.DefaultAppFile()); err != nil {
		return err
	}

	for _, path := range []string{t.RequirementsFile} {
		if err = createFile(filepath.Join(basePath, path)); err != nil {
			return err
		}
	}
	if err := bootstrapRequirements(t, basePath); err != nil {
		return err
	}
	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, t.CoverImage)); err != nil {
		return err
	}

	if err = CreateAndWriteIfFileNotExist(filepath.Join(basePath, manifest.ManifestPath), manifestToml); err != nil {
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
