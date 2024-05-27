package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipFolder compresses the current directory into a zip-file.
// It returns an error if anything fails, else nil.
func ZipFolder(zipFile *os.File, exclude []string) error {
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldExclude(exclude, path) || info.Name() == zipFile.Name() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Ensure the header name is a relative path to avoid file path disclosure.
		relPath, err := filepath.Rel(".", path)
		if err != nil {
			return err
		}
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

func shouldExclude(excludedPatterns []string, path string) bool {
	for _, pattern := range excludedPatterns {
		if Match(pattern, path) {
			return true
		}
	}

	return false
}
