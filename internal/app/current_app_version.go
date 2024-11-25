package app

import (
	"context"
	"errors"

	"github.com/hasura/go-graphql-client"
)

var ErrNotDeployed = errors.New("app is not deployed")

type CurrentAppVersionInput struct {
	OrganizationSlug string
	AppSlug          string
}

type CurrentAppVersionOutput struct {
	AppVersionID string
}

const queryCurrentAppVersionText = `
query CLIReadCurrentAppVersion($orgSlug: String!, $appSlug: String!) {
	app(organizationSlug: $orgSlug, appSlug: $appSlug) {
		defaultDeployment {
			current {
				appVersion {
					id
				}
			}
		}
	}
}
`

type currentAppVersionResponse struct {
	App struct {
		DefaultDeployment *struct {
			Current struct {
				AppVersion struct {
					ID string
				}
			}
		}
	}
}

func (s *Service) CurrentAppVersion(ctx context.Context, input CurrentAppVersionInput) (CurrentAppVersionOutput, error) {
	var resp currentAppVersionResponse

	variables := map[string]any{"orgSlug": input.OrganizationSlug, "appSlug": input.AppSlug}
	err := s.client.Exec(ctx, queryCurrentAppVersionText, &resp, variables, graphql.OperationName("CLIReadCurrentAppVersion"))
	if err != nil {
		return CurrentAppVersionOutput{}, convertErrors(err)
	}

	if resp.App.DefaultDeployment == nil {
		return CurrentAppVersionOutput{}, ErrNotDeployed
	}

	return CurrentAppVersionOutput{AppVersionID: resp.App.DefaultDeployment.Current.AppVersion.ID}, nil
}
