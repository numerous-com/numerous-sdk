package organization

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type organizationCreateResponse struct {
	OrganizationCreate Organization
}

func Create(name string, client *gqlclient.Client) (Organization, error) {
	resp := organizationCreateResponse{}
	op := createOrganizationCreateOperation(name)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		if err.Error() == ErrOrganizationNameInvalidCharacter.Error() {
			return resp.OrganizationCreate, ErrOrganizationNameInvalidCharacter
		}

		return resp.OrganizationCreate, err
	}

	return resp.OrganizationCreate, nil
}

func createOrganizationCreateOperation(name string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	mutation OrganizationCreate($name: String!) {
		organizationCreate(input: { name: $name }) {
			id
			name
			slug
		}
	}
`)

	op.Var("name", name)

	return op
}
