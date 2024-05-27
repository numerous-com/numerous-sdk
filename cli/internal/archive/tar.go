package archive

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const filePermission = 0o755

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

		if shouldExclude(exclude, relPath) {
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

func TarExtract(content io.Reader, dest string) error {
	tr := tar.NewReader(content)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, filePermission); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
