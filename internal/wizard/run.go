package wizard

import (
	"errors"
	"fmt"

	"numerous.com/cli/internal/manifest"
)

var ErrStopInit = errors.New("stop app init")

func questionsNeeded(params RunWizardParams) bool {
	if params.App.Name == "" {
		return true
	}

	if params.Python.Library.Key != "" && params.Python.Library.Key != manifest.DockerfileLibraryKey && params.Python.AppFile != "" && params.Python.RequirementsFile != "" {
		return false
	}

	if params.App.LibraryKey == manifest.DockerfileLibraryKey && params.Docker.Dockerfile != "" && params.Docker.Context != "" && params.Docker.Port != 0 {
		return false
	}

	return true
}

type RunWizardParams struct {
	AppDir string
	App    AppAnswers
	Python PythonAnswers
	Docker DockerAnswers
}

func Run(asker Asker, params RunWizardParams) (*manifest.Manifest, error) {
	python := params.Python.ToManifest()
	docker := params.Docker.ToManifest()

	if !questionsNeeded(params) {
		var port uint
		if python != nil {
			port = python.Library.Port
		} else if params.Docker.Port != 0 {
			port = uint(params.Docker.Port)
		}

		return &manifest.Manifest{
			App:    params.App.ToManifestApp(port),
			Python: python,
			Docker: docker,
		}, nil
	}

	fmt.Println("Hi there, welcome to Numerous.")
	fmt.Println("We're happy you're here!")
	fmt.Println("Let's get started by entering basic information about your app.")

	continueWizard, err := UseOrCreateAppFolder(asker, params.AppDir)
	if err != nil {
		return nil, err
	} else if !continueWizard {
		return nil, ErrStopInit
	}

	appAnswers, err := appWizard(asker, params.App)
	if err != nil {
		return nil, err
	}

	m := manifest.Manifest{}

	var port uint
	if appAnswers.LibraryName == dockerfileLibraryName || appAnswers.LibraryKey == manifest.DockerfileLibraryKey {
		dockerAnswers, err := dockerWizard(asker, params.Docker)
		if err != nil {
			return nil, err
		}
		m.Docker = dockerAnswers.ToManifest()
		port = uint(dockerAnswers.Port)
	} else {
		pythonAnswers, err := pythonWizard(asker, appAnswers.LibraryName, params.Python)
		if err != nil {
			return nil, err
		}
		m.Python = pythonAnswers.ToManifest()
		port = m.Python.Library.Port
	}

	m.App = appAnswers.ToManifestApp(port)

	return &m, nil
}
