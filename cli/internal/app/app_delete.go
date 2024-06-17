package app

import (
	"context"

	"numerous/cli/internal/gql"
)

const deleteMutation string = `
	mutation DeleteApp($slug: String!, $name: String!) {
		appDelete(input: {organizationSlug: $slug, appName: $name}) {
			__typename
		}
	}
`

type deleteAppResponse struct {
	AppDelete struct {
		Typename string `graphql:"__typename"`
	}
}

type DeleteAppInput struct {
	OrganizationSlug string
	Name             string
}

func (s *Service) Delete(ctx context.Context, input DeleteAppInput) error {
	resp := deleteAppResponse{}
	vars := map[string]any{"slug": input.OrganizationSlug, "name": input.Name}

	err := s.client.Exec(ctx, deleteMutation, &resp, vars)
	if err != nil {
		return gql.CheckAccessDenied(err)
	}

	return nil
}
