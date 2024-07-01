package app

import (
	"context"
)

type CreateAppInput struct {
	OrganizationSlug string
	AppSlug          string
	DisplayName      string
	Description      string
}

type CreateAppOutput struct {
	AppID string
}

const appCreateText = `
mutation AppCreate($orgSlug: String!, $appSlug: String!, $displayName: String!, $description: String!) {
	appCreate(organizationSlug: $orgSlug, appData: {appSlug: $appSlug, displayName: $displayName, description: $description}) {
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
		"orgSlug":     input.OrganizationSlug,
		"appSlug":     input.AppSlug,
		"displayName": input.DisplayName,
		"description": input.Description,
	}
	err := s.client.Exec(ctx, appCreateText, &resp, variables)
	if err != nil {
		return CreateAppOutput{}, ConvertErrors(err)
	}

	return CreateAppOutput{
		AppID: resp.AppCreate.ID,
	}, nil
}
