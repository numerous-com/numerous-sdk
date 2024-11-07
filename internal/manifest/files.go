package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"
)

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// Creates file if it does not exists
func createFile(path string) error {
	if exists, err := fileExists(path); err != nil {
		return err
	} else if !exists {
		// file does not exist, create it
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	return nil
}

func createAndWriteIfFileNotExist(path string, content string) error {
	path = filepath.Clean(path)
	_, err := os.Stat(path)
	if err == nil {
		fmt.Printf("Skipping creation of %q; it already exists\n", path)
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		output.PrintErrorDetails("Could not write to %q", err, path)
	}

	return nil
}

// Writes content to a specific path
func writeOrAppendFile(path string, content string) error {
	path = filepath.Clean(path)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		output.PrintErrorDetails("Could not open %q", err, path)
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		output.PrintErrorDetails("Could not determine the file size of %q", err, path)
		return err
	} else if fileStat.Size() != 0 {
		content = "\n" + content
	}

	_, err = file.WriteString(content)
	if err != nil {
		output.PrintErrorDetails("Could not write to %q", err, path)
	}

	return nil
}

// Generates and creates file containing the tools id
func createAppIDFile(path string, id string) error {
	appIDFile := filepath.Join(path, dir.AppIDFileName)
	if err := createFile(appIDFile); err != nil {
		output.PrintUnknownError(err)
		return err
	}

	return writeOrAppendFile(appIDFile, id)
}

// Creates and adds the item to .gitignore
func addToGitIgnore(path string, toIgnore []string) error {
	gitignorePath := filepath.Join(path, ".gitignore")
	if err := createFile(gitignorePath); err != nil {
		output.PrintUnknownError(err)
		return err
	}

	return writeOrAppendFile(gitignorePath, strings.Join(toIgnore, "\n")+"\n")
}
