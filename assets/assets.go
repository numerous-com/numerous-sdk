package assets

import (
	"embed"
	"io"
	"os"
)

//go:embed images/placeholder_tool_cover.png
var image embed.FS

func CopyToolPlaceholderCover(destPath string) error {
	srcFile, err := image.Open("images/placeholder_tool_cover.png")
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
