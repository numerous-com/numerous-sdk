package manifest

import (
	"errors"
	"path/filepath"

	"numerous.com/cli/assets"
	"numerous.com/cli/internal/dir"
)

const EnvFileName string = ".env"

var ErrEncodingManifest = errors.New("error encoding manifest")

func (m *Manifest) BootstrapFiles(toolID string, basePath string) error {
	manifestToml, err := m.ToTOML()
	if err != nil {
		return ErrEncodingManifest
	}

	if toolID != "" {
		if err = createAppIDFile(basePath, toolID); err != nil {
			return err
		}
	}

	gitignoreLines := []string{"# added by numerous init\n", EnvFileName}
	if toolID != "" {
		gitignoreLines = append(gitignoreLines, dir.AppIDFileName)
	}

	if err := addToGitIgnore(basePath, gitignoreLines); err != nil {
		return err
	}

	if m.Python != nil {
		if err := m.Python.bootstrapFiles(basePath); err != nil {
			return err
		}
	} else if m.Docker != nil {
		if err := m.Docker.bootstrapFiles(basePath); err != nil {
			return err
		}
	}

	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, m.CoverImage)); err != nil {
		return err
	}

	if err = createAndWriteIfFileNotExist(filepath.Join(basePath, ManifestPath), manifestToml); err != nil {
		return err
	}

	return nil
}
