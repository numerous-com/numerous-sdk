package organization

import (
	"errors"
	"fmt"
)

var ErrOrganizationNameInvalidCharacter = errors.New("gqlclient: server failure: organization name contains invalid characters")

func (o Organization) String() string {
	return fmt.Sprintf(`
Organization:
  name     %s
  url      %s
	`, o.Name, "https://numerous.com/app/organization/"+o.Slug)
}
