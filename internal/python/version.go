package python

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

var (
	ErrDetectPythonExecutable = errors.New("could not detect python executable")
	ErrDetectPythonVersion    = errors.New("could not detect python version")
)

// Returns the python version used in the environment, or 3.11 as a fallback if
// it cannot be detected.
func PythonVersion() string {
	fallbackVersion := "3.11"

	version, err := getPythonVersion()

	if err == nil {
		return version
	}

	if errors.Is(err, ErrDetectPythonExecutable) {
		fmt.Printf("Python interpeter not found, setting Python version to '%s' for the app.\n", fallbackVersion)
	}

	if errors.Is(err, ErrDetectPythonVersion) {
		fmt.Printf("Could not parse python version '%s', setting Python version to '%s' for the app.\n", version, fallbackVersion)
	}

	return fallbackVersion
}

func getPythonVersion() (string, error) {
	p, err := execPythonVersionCommand()
	if err != nil {
		return "", err
	}

	version, err := extractPythonVersion(p)
	if err != nil {
		return "", err
	}

	return version, nil
}

func execPythonVersionCommand() ([]byte, error) {
	pythonExes := []string{"python", "python3", "python3.9", "python3.10", "python3.11", "python3.12"}

	for _, pythonExe := range pythonExes {
		if p, err := exec.Command(pythonExe, "-V").Output(); err == nil {
			return p, nil
		}
	}

	return []byte{}, ErrDetectPythonExecutable
}

func extractPythonVersion(p []byte) (string, error) {
	verifyPythonOutput := regexp.MustCompile(`Python [\d]+.[\d]+\+?`)
	if !verifyPythonOutput.Match(p) {
		return "", ErrDetectPythonVersion
	}

	getVersionRegex := regexp.MustCompile(`[\d]+.[\d]+`)
	version := getVersionRegex.Find(p)

	return string(version), nil
}
