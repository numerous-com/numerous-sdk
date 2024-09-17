package init

import (
	"errors"
	"fmt"
	"strings"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/python"
	"numerous.com/cli/internal/wizard"
)

type InitializeParams struct {
	Name             string
	Desc             string
	AppDir           string
	LibraryKey       string
	AppFile          string
	RequirementsFile string
	Dockerfile       string
	DockerContext    string
}

func Initialize(asker wizard.Asker, params InitializeParams) (*manifest.Manifest, error) {
	fmt.Printf("params=%v\n", params)
	if exist, _ := manifest.ManifestExists(params.AppDir); exist {
		output.PrintError(
			"An app is already initialized in \"%s\"",
			"💡 You can initialize an app in another folder by specifying a\n"+
				"   path in the command, like below:\n\n"+
				"numerous init ./my-app-folder\n\n",
			params.AppDir,
		)

		return nil, ErrAppAlreadyInitialized
	}

	lib, err := manifest.GetLibraryByKey(params.LibraryKey)
	if params.LibraryKey != "" && params.LibraryKey != manifest.DockerfileLibraryKey && errors.Is(err, manifest.ErrUnsupportedLibrary) {
		output.PrintError(
			"Unsupported library",
			"The specified library %s is not supported. Supported libraries are %s.",
			params.LibraryKey,
			manifest.SupportedLibraryValuesList(),
		)

		return nil, err
	}

	runWizardParams := wizard.RunWizardParams{
		ProjectFolderPath: params.AppDir,
		App: wizard.AppAnswers{
			Name:        params.Name,
			Description: params.Desc,
			LibraryKey:  params.LibraryKey,
			LibraryName: lib.Name,
		},
		Python: wizard.PythonAnswers{
			RequirementsFile: params.RequirementsFile,
			AppFile:          params.AppFile,
			Library:          lib,
		},
		Docker: wizard.DockerAnswers{
			Dockerfile: params.Dockerfile,
			Context:    params.DockerContext,
		},
	}
	m, err := wizard.Run(asker, runWizardParams)
	if err != nil {
		output.PrintErrorDetails("Error running initialization wizard", err)
		return nil, err
	}

	if m.Python != nil && m.Python.Library.Key != "" {
		m.Python.Version = python.PythonVersion()
	}

	err = m.BootstrapFiles("", params.AppDir)
	switch {
	case err == manifest.ErrEncodingManifest:
		output.PrintErrorDetails("Error encoding manifest file", err)
	case err != nil:
		output.PrintErrorDetails("Error bootstrapping files.", err)
		return nil, err
	}

	return m, nil
}

func PathArgumentHandler(providedPath string, currentPath string) string {
	appPath := providedPath
	if providedPath != "." {
		pathBegin := string([]rune(providedPath)[0:2])
		if pathBegin == "./" || pathBegin == ".\\" {
			appPath = strings.Replace(appPath, ".", currentPath, 1)
		} else {
			appPath = providedPath
		}
	} else {
		appPath = currentPath
	}

	return appPath
}