package dev

import (
	"errors"
	"fmt"
	"strings"

	"numerous/cli/appdevsession"

	"github.com/spf13/cobra"
)

type AppLocation struct {
	ModulePath string
	ClassName  string
}

var (
	port              string
	pythonInterpreter string
	appLocation       AppLocation
)

var DevCmd = &cobra.Command{
	Use:   "dev MODULE:CLASS",
	Run:   dev,
	Short: "Develop and run numerous app engine apps locally.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("an app module and class must be specified")
		} else if len(args) > 1 {
			return errors.New("too many arguments. only one app can be specified")
		}

		parsedAppLocation, err := parseAppLocation(args[0])
		if err != nil {
			return err
		}

		appLocation = parsedAppLocation

		return nil
	},
}

func parseAppLocation(s string) (AppLocation, error) {
	var location AppLocation
	parts := strings.Split(s, ":")
	allowedNumberOfParts := 2
	if len(parts) != allowedNumberOfParts {
		return location, fmt.Errorf("invalid app location '%s', there must be exactly 1 ':'", s)
	} else {
		location.ModulePath = parts[0]
		location.ClassName = parts[1]

		return location, nil
	}
}

func dev(cmd *cobra.Command, args []string) {
	appdevsession.CreateAndRunDevSession(pythonInterpreter, appLocation.ModulePath, appLocation.ClassName, port)
}

func init() {
	flags := DevCmd.Flags()
	flags.StringVar(&port, "port", "7001", "The GraphQL Port")
	flags.StringVar(
		&pythonInterpreter,
		"python",
		"python",
		"Path to the python interpreter, eg. \"python\", or \"./venv/bin/python\"",
	)
}
