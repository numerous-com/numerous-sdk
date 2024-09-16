package manifest

import (
	"errors"
	"path/filepath"

	"numerous.com/cli/assets"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"
)

const EnvFileName string = ".env"

var (
	ErrEncodingManifest      = errors.New("error encoding manifest")
	ErrNoLibraryBootstrapped = errors.New("no library bootstrapped")
)

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

	switch {
	case m.Python != nil && m.Python.Library.Key != "":
		if err := m.Python.bootstrapFiles(basePath); err != nil {
			return err
		}
	case m.Docker != nil:
		err := m.Docker.bootstrapFiles(basePath)
		if errors.Is(err, ErrNoBootstrapDockerfileExists) {
			output.Notify(
				"Skipping Dockerfile bootstrapping",
				"A Dockerfile already exists at %q, so Numerous CLI will not bootstrap an example Dockerfile app.",
				filepath.Join(basePath, m.Docker.Dockerfile),
			)
		} else if err != nil {
			return err
		}
	default:
		return ErrNoLibraryBootstrapped
	}

	if err = assets.CopyToolPlaceholderCover(filepath.Join(basePath, m.CoverImage)); err != nil {
		return err
	}

	if err = createAndWriteIfFileNotExist(filepath.Join(basePath, ManifestPath), manifestToml); err != nil {
		return err
	}

	return nil
}
