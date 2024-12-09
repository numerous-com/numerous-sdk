package args

import (
	"github.com/spf13/pflag"
)

type AppIdentifierArg struct {
	OrganizationSlug string
	AppSlug          string
}

func (a *AppIdentifierArg) AddAppIdentifierFlags(flags *pflag.FlagSet, action string) {
	AddOrganizationSlugFlag(flags, action, &a.OrganizationSlug)
	flags.StringVarP(&a.AppSlug, "app", "a", "", "An app slug identifier of the app "+action+".")
}

func AddOrganizationSlugFlag(flags *pflag.FlagSet, action string, orgSlug *string) {
	flags.StringVarP(orgSlug, "organization", "o", "", "The organization slug identifier of the app "+action+". List available organizations with 'numerous organization list'.")
}
