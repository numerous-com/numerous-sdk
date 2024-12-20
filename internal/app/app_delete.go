package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

const deleteMutation string = `
	mutation CLIAppDelete($orgSlug: String!, $appSlug: String!) {
		appDelete(input: {organizationSlug: $orgSlug, appSlug: $appSlug}) {
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
	AppSlug          string
}

func (s *Service) Delete(ctx context.Context, input DeleteAppInput) error {
	resp := deleteAppResponse{}
	vars := map[string]any{"orgSlug": input.OrganizationSlug, "appSlug": input.AppSlug}

	err := s.client.Exec(ctx, deleteMutation, &resp, vars, graphql.OperationName("CLIAppDelete"))
	if err != nil {
		return convertErrors(err)
	}

	return nil
}
