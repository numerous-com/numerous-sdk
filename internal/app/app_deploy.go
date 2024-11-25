package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"numerous.com/cli/internal/gql/secret"
)

type DeployAppInput struct {
	AppVersionID    string
	AppRelativePath string
	Secrets         map[string]string
}

type DeployAppOutput struct {
	DeploymentVersionID string
}

const appDeployText = `
mutation CLIAppDeploy($appVersionID: ID!, $secrets: [AppSecret!], $appRelativePath: String!) {
	appDeploy(appVersionID: $appVersionID, input: {appRelativePath: $appRelativePath, secrets: $secrets}) {
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
	variables := map[string]any{
		"appVersionID":    input.AppVersionID,
		"secrets":         convertedSecrets,
		"appRelativePath": input.AppRelativePath,
	}

	err := s.client.Exec(ctx, appDeployText, &resp, variables, graphql.OperationName("CLIAppDeploy"))
	if err != nil {
		return DeployAppOutput{}, convertErrors(err)
	}

	return DeployAppOutput{DeploymentVersionID: resp.AppDeploy.ID}, nil
}
