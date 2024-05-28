package app

import "context"

type DeployAppInput struct {
	AppVersionID string
}

type DeployAppOutput struct {
	DeploymentVersionID string
}

const appDeployText = `
mutation AppDeploy($appVersionID: ID!) {
	appDeploy(appVersionID: $appVersionID) {
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
	variables := map[string]any{"appVersionID": input.AppVersionID}

	err := s.client.Exec(ctx, appDeployText, &resp, variables)
	if err != nil {
		return DeployAppOutput{}, err
	}

	return DeployAppOutput{DeploymentVersionID: resp.AppDeploy.ID}, nil
}
