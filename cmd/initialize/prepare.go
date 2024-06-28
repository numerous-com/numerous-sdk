package initialize

import (
	"errors"
	"os"
	"strings"

	"numerous.com/cli/cmd/initialize/wizard"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/python"
)

func PrepareInit(args []string) (string, *manifest.Manifest, error) {
	appDir, err := os.Getwd()
	if err != nil {
		return "", nil, ErrGetWorkDir
	}

	if len(args) != 0 {
		appDir = PathArgumentHandler(args[0], appDir)
	}

	if exist, _ := manifest.ManifestExists(appDir); exist {
		output.PrintError(
			"An app is already initialized in \"%s\"",
			"💡 You can initialize an app in another folder by specifying a\n"+
				"   path in the command, like below:\n\n"+
				"numerous init ./my-app-folder\n\n",
			appDir,
		)

		return "", nil, ErrAppAlreadyInitialized
	}

	lib, err := manifest.GetLibraryByKey(libraryKey)
	if libraryKey != "" && errors.Is(err, manifest.ErrUnsupportedLibrary) {
		output.PrintError(
			"Unsupported library",
			"The specified library %s is not supported. Supported libraries are %s.",
			libraryKey,
			manifest.SupportedLibraryValuesList(),
		)

		return "", nil, err
	} else if libraryKey != "" && err != nil {
		output.PrintErrorDetails("Unsupported library", err)
	}

	pythonVersion := python.PythonVersion()

	m := manifest.New(lib, name, desc, pythonVersion, appFile, requirementsFile)
	if continueBootstrap, err := wizard.RunInitAppWizard(appDir, m); err != nil {
		output.PrintErrorDetails("Error running initialization wizard", err)
		return "", nil, err
	} else if !continueBootstrap {
		return "", nil, ErrStopBootstrap
	}

	return appDir, m, nil
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
