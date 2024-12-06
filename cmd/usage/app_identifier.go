package usage

import "fmt"

const appIdentifierFormat string = `The app %s is specified firstly by --app and
--organization flags, secondly by the deployment section in the manifest
(numerous.toml), and finally by the local configuration (see "numerous config").

If no app is identified either by --app or by the default deployment section in
the manifest, the app name in the manifest is converted into a slug, e.g.
"My Awesome App" would become "my-awesome-app".

App and organization identifiers must contain only lower-case alphanumeric
characters and dashes.`

func AppIdentifier(action string) string {
	return fmt.Sprintf(appIdentifierFormat, action)
}
