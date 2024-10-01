package manifest

import (
	"os"
	"path/filepath"

	"numerous.com/cli/internal/requirements"
)

func (p Python) bootstrapFiles(basePath string) error {
	appFilePath := filepath.Join(basePath, p.AppFile)
	if err := createAndWriteIfFileNotExist(appFilePath, p.Library.DefaultAppFile()); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(basePath, p.RequirementsFile)
	if err := createFile(requirementsFilePath); err != nil {
		return err
	}

	if err := p.bootstrapRequirements(basePath); err != nil {
		return err
	}

	return nil
}

func (p Python) bootstrapRequirements(basePath string) error {
	requirementsPath := filepath.Join(basePath, p.RequirementsFile)
	rfile, err := os.Open(requirementsPath)
	if err != nil {
		return err
	}

	req, err := requirements.Read(rfile)
	if err != nil {
		return err
	}

	for _, requirement := range p.Library.Requirements {
		req.Add(requirement)
	}

	wfile, err := os.Create(requirementsPath)
	if err != nil {
		return err
	}
	defer wfile.Close()

	return req.Write(wfile)
}
