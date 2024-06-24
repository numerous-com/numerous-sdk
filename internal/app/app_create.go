package app

import (
	"context"

	"numerous.com/cli/internal/gql"
)

type CreateAppInput struct {
	OrganizationSlug string
	Name             string
	DisplayName      string
	Description      string
}

type CreateAppOutput struct {
	AppID string
}

const appCreateText = `
mutation AppCreate($slug: String!, $name: String!, $displayName: String!, $description: String!) {
	appCreate(organizationSlug: $slug, appData: {name: $name, displayName: $displayName, description: $description}) {
		id
	}
}
`

type appCreateResponse struct {
	AppCreate struct {
		ID string
	}
}

func (s *Service) Create(ctx context.Context, input CreateAppInput) (CreateAppOutput, error) {
	var resp appCreateResponse

	variables := map[string]any{
		"slug":        input.OrganizationSlug,
		"name":        input.Name,
		"displayName": input.DisplayName,
		"description": input.Description,
	}
	err := s.client.Exec(ctx, appCreateText, &resp, variables)
	if err != nil {
		return CreateAppOutput{}, gql.CheckAccessDenied(err)
	}

	return CreateAppOutput{
		AppID: resp.AppCreate.ID,
	}, nil
}
