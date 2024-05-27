package archive

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TarCreate creates a tar file at `destPath`, from the given `srcDir`,
// excluding files matching patterns in `exclude`.
func TarCreate(srcDir string, destPath string, exclude []string) error {
	tarFile, err := os.Create(destPath)
	if err != nil {
		return err
	}

	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	err = filepath.Walk(srcDir, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, fileName)
		if err != nil {
			return err
		}

		if shouldExclude(exclude, relPath) || tarFile.Name() == path.Join(srcDir, fi.Name()) {
			return nil
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		header.Name = strings.ReplaceAll(relPath, "\\", "/")

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Ignore all non-regular files (e.g. directories, links, executables, etc.)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// Copy regular files
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return err
		}

		return nil
	})

	return err
}
