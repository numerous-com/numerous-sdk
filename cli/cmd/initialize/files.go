package initialize

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"numerous/cli/tool"
)

// Creates file if it does not exists
func createFile(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		// file does not exist, create it
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		return err
	}

	return nil
}

func CreateAndWriteIfFileNotExist(path string, content string) error {
	_, err := os.Stat(path)
	if err == nil {
		fmt.Printf("Skipping creation of app file: %s already exists\n", path)
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

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Could not write to %s\n", path)
	}

	return nil
}

// Writes content to a specific path
func writeOrAppendFile(path string, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Printf("Could not open '%s'\n", path)
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Printf("Could not determine file size of '%s'\n", path)
		return err
	} else if fileStat.Size() != 0 {
		content = "\n" + content
	}

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Could not write to '%s'\n", path)
	}

	return nil
}

// Generates and creates file containing the tools id
func createToolIDFile(path string, id string) error {
	toolFile := filepath.Join(path, tool.ToolIDFileName)
	if err := createFile(toolFile); err != nil {
		fmt.Printf("Error creating tool id file\nError: %s", err)
		return err
	}

	return writeOrAppendFile(toolFile, id)
}

// Creates and adds the item to .gitignore
func addToGitIgnore(path string, toIgnore string) error {
	gitignorePath := filepath.Join(path, ".gitignore")
	if err := createFile(gitignorePath); err != nil {
		fmt.Println("Error creating .gitignore")
		return err
	}

	return writeOrAppendFile(gitignorePath, toIgnore)
}
