package manifest

import (
	"os"
	"strings"

	"numerous.com/cli/cmd/output"
)

// Validates that the given app file is valid for this library. Returns false,
// if the app is in a state, where it does not make sense, to be able to push
// the app.
func (m *Manifest) ValidateApp() (bool, error) {
	return m.Library.ValidateApp(m.AppFile)
}

func (l Library) ValidateApp(appFile string) (bool, error) {
	switch l.Key {
	case "numerous":
		return validateNumerousApp(appFile)
	default:
		return true, nil
	}
}

func validateNumerousApp(appFile string) (bool, error) {
	data, err := os.ReadFile(appFile)
	if err != nil {
		return false, err
	}

	filecontent := string(data)
	if strings.Contains(filecontent, "appdef =") || strings.Contains(filecontent, "class appdef") {
		return true, nil
	} else {
		output.PrintError("Your app file must have an app definition called 'appdef'", strings.Join(
			[]string{
				"You can solve this by assigning your app definition to this name, for example:",
				"",
				"@app",
				"class MyApp:",
				"    my_field: str",
				"",
				"appdef = MyApp",
			}, "\n"),
		)

		return false, nil
	}
}
