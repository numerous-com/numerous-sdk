package organization

import (
	"errors"
	"fmt"

	"numerous.com/cli/internal/links"
)

var ErrOrganizationNameInvalidCharacter = errors.New("gqlclient: server failure: organization name contains invalid characters")

func (o Organization) String() string {
	return fmt.Sprintf(`
Organization:
  name     %s
  url      %s
	`, o.Name, links.GetOrganizationURL(o.Slug))
}
