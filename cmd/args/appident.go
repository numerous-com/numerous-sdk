package args

import (
	"github.com/spf13/pflag"
)

type AppIdentifierArg struct {
	OrganizationSlug string
	AppSlug          string
}

func (a *AppIdentifierArg) AddAppIdentifierFlags(flags *pflag.FlagSet, action string) {
	flags.StringVarP(&a.OrganizationSlug, "organization", "o", "", "The organization slug identifier of the app "+action+". List available organizations with 'numerous organization list'.")
	flags.StringVarP(&a.AppSlug, "app", "a", "", "An app slug identifier of the app "+action+".")
}
