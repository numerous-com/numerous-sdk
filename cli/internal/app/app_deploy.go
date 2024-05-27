package app

import (
	"context"

	"numerous/cli/internal/gql/secret"
)

type DeployAppInput struct {
	AppVersionID string
	Secrets      map[string]string
}

type DeployAppOutput struct {
	DeploymentVersionID string
}

const appDeployText = `
mutation AppDeploy($appVersionID: ID!, $secrets: [AppSecret!]) {
	appDeploy(appVersionID: $appVersionID, input: {secrets: $secrets}) {
		id
	}
}
`

type appDeployResponse struct {
	AppDeploy struct {
		ID string
	}
}

func (s *Service) DeployApp(ctx context.Context, input DeployAppInput) (DeployAppOutput, error) {
	var resp appDeployResponse
	convertedSecrets := secret.AppSecretsFromMap(input.Secrets)
	variables := map[string]any{"appVersionID": input.AppVersionID, "secrets": convertedSecrets}

	err := s.client.Exec(ctx, appDeployText, &resp, variables)
	if err != nil {
		return DeployAppOutput{}, err
	}

	return DeployAppOutput{DeploymentVersionID: resp.AppDeploy.ID}, nil
}
