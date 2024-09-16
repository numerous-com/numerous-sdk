package dir

import (
	"errors"
	"os"
	"path/filepath"

	"numerous.com/cli/cmd/output"
)

const (
	ToolIDFileName string = ".tool_id.txt" // supported for backwards compatibility
	AppIDFileName  string = ".app_id.txt"
)

var ErrAppIDNotFound = errors.New("app id not found")

func AppIDExists(dir string) (bool, error) {
	appIDFilePath := filepath.Join(dir, AppIDFileName)
	_, err := os.Stat(appIDFilePath)
	if err == nil {
		return true, nil
	} else if !errors.Is(err, os.ErrNotExist) && err.Error() != "no such file or directory" {
		return true, err
	}

	toolIDFilePath := filepath.Join(dir, ToolIDFileName)
	_, err = os.Stat(toolIDFilePath)
	if err == nil {
		return true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
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

func PrintReadAppIDErrors(err error, appDir string) {
	if err == ErrAppIDNotFound {
		output.PrintErrorAppNotInitialized(appDir)
	} else if err != nil {
		output.PrintErrorDetails("An error occurred reading the app ID", err)
	}
}
