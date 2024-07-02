package manifest

import (
	"errors"
	"os"
	"strings"
)

var ErrValidateNumerousApp = errors.New("error validating numerous app engine")

// Validates that the given app file is valid for this library. Returns false,
// if the app is in a state where it does not make sense to be able to push
// the app.
func (m *Manifest) ValidateApp() error {
	return m.Library.ValidateApp(m.AppFile)
}

func (l Library) ValidateApp(appFile string) error {
	switch l.Key {
	case "numerous":
		return validateNumerousApp(appFile)
	default:
		return nil
	}
}

func validateNumerousApp(appFile string) error {
	data, err := os.ReadFile(appFile)
	if err != nil {
		return err
	}

	filecontent := string(data)
	if strings.Contains(filecontent, "appdef =") || strings.Contains(filecontent, "class appdef") {
		return nil
	} else {
		return ErrValidateNumerousApp
	}
}
