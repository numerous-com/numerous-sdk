package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type ReadAppInput struct {
	OrganizationSlug string
	AppSlug          string
}

type ReadAppOutput struct {
	AppID          string
	AppDisplayName string
	AppDescription string
}

const queryAppText = `
query CLIAppRead($orgSlug: String!, $appSlug: String!) {
	app(organizationSlug: $orgSlug, appSlug: $appSlug) {
		id
		displayName
		description
	}
}
`

type appResponse struct {
	App struct {
		ID          string
		DisplayName string
		Description string
	}
}

func (s *Service) ReadApp(ctx context.Context, input ReadAppInput) (ReadAppOutput, error) {
	var resp appResponse

	variables := map[string]any{"orgSlug": input.OrganizationSlug, "appSlug": input.AppSlug}
	err := s.client.Exec(ctx, queryAppText, &resp, variables, graphql.OperationName("CLIAppRead"))
	if err == nil {
		return ReadAppOutput{AppID: resp.App.ID, AppDisplayName: resp.App.DisplayName, AppDescription: resp.App.Description}, nil
	}

	return ReadAppOutput{}, convertErrors(err)
}
