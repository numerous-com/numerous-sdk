package app

import (
	"context"
)

type CreateAppVersionInput struct {
	AppID string
}

type CreateAppVersionOutput struct {
	AppVersionID string
}

const appVersionCreateText = `
mutation AppVersionCreate($appID: ID!) {
	appVersionCreate(appID: $appID) {
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
		"appID": input.AppID,
	}
	err := s.client.Exec(ctx, appVersionCreateText, &resp, variables)
	if err != nil {
		return CreateAppVersionOutput{}, err
	}

	return CreateAppVersionOutput{
		AppVersionID: resp.AppVersionCreate.ID,
	}, nil
}
