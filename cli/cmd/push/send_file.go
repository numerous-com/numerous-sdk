package push

import (
	"io/fs"
	"os"

	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/build"
)

func uploadZipFile(zipFilePath string, appID string) (string, error) {
	var filePermission fs.FileMode = 0o666
	zipFile, err := os.OpenFile(zipFilePath, os.O_CREATE|os.O_RDWR, filePermission)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	build, err := build.Push(zipFile, appID, gql.GetClient())
	if err != nil {
		return "", err
	}

	return build.BuildID, nil
}
