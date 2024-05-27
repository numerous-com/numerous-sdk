package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipCreate compresses the given source directory into a zip-file.
// It returns an error if anything fails, else nil.
func ZipCreate(srcDir, destPath string, exclude []string) error {
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if shouldExclude(exclude, relPath) || info.Name() == zipFile.Name() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Ensure the header name is a relative path to avoid file path disclosure.
		header.Name = relPath

		// Ensure folder names end with a slash to distinguish them in the zip.
		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Only copy file content; directories are added as empty entries.
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)

			return err // err could be nil or an actual error
		}

		return nil
	})
}
