package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

const queryAppDeploymentIDText = `
query CLIAppDeploymentID($orgSlug: String!, $appSlug: String!) {
	app(organizationSlug: $orgSlug, appSlug: $appSlug) {
		id
		defaultDeployment {
			id
		}
	}
}
`

type appDeploymentResponse struct {
	App struct {
		ID                string
		DefaultDeployment *struct {
			ID string
		}
	}
}

func (s *Service) GetAppDeploymentID(ctx context.Context, organizationSlug, appSlug string) (string, error) {
	var resp appDeploymentResponse

	variables := map[string]any{"orgSlug": organizationSlug, "appSlug": appSlug}
	err := s.client.Exec(ctx, queryAppDeploymentIDText, &resp, variables, graphql.OperationName("CLIAppDeploymentID"))
	if err != nil {
		return "", convertErrors(err)
	}

	if resp.App.ID == "" {
		return "", ErrAppNotFound
	}

	if resp.App.DefaultDeployment == nil {
		return "", ErrDeploymentNotFound
	}

	return resp.App.DefaultDeployment.ID, nil
}
