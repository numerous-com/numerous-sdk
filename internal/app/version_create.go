package app

import (
	"context"
)

type CreateAppVersionInput struct {
	AppID   string
	Version string
	Message string
}

type CreateAppVersionOutput struct {
	AppVersionID string
}

const appVersionCreateText = `
mutation AppVersionCreate($appID: ID!, $version: String, $message: String!) {
	appVersionCreate(appID: $appID, input: {version: $version, message: $message}) {
		id
	}
}
`

type appVersionCreateResponse struct {
	AppVersionCreate struct {
		ID string
	}
}

func (s *Service) CreateVersion(ctx context.Context, input CreateAppVersionInput) (CreateAppVersionOutput, error) {
	var resp appVersionCreateResponse

	variables := map[string]any{
		"appID":   input.AppID,
		"message": input.Message,
	}

	if input.Version != "" {
		variables["version"] = input.Version
	}

	err := s.client.Exec(ctx, appVersionCreateText, &resp, variables)
	if err != nil {
		return CreateAppVersionOutput{}, err
	}

	return CreateAppVersionOutput{
		AppVersionID: resp.AppVersionCreate.ID,
	}, nil
}
