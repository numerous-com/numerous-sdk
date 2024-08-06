package push

import (
	"path/filepath"
	"strings"
)

func CheckAndReturnSubpath(basePath, subPath string) (bool, string, error) {
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return false, "", err
	}

	absSubPath, err := filepath.Abs(subPath)
	if err != nil {
		return false, "", err
	}

	appPath, err := filepath.Rel(absBasePath, absSubPath)
	if err != nil {
		return false, "", err
	}

	isSub := !strings.HasPrefix(appPath, "..") && !strings.HasPrefix(appPath, string(filepath.Separator))

	return isSub, appPath, nil
}
