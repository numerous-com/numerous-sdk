package app

import (
	"context"
)

type AppVersionUploadURLInput struct {
	AppVersionID string
}

type AppVersionUploadURLOutput struct {
	UploadURL string
}

const appVersionUploadURLText = `
mutation AppVersionUploadURL($appVersionID: ID!) {
	appVersionUploadURL(appVersionID: $appVersionID) {
		url
	}
}
`

type appVersionUploadURLResponse struct {
	AppVersionUploadURL struct {
		URL string
	}
}

func (s *Service) AppVersionUploadURL(ctx context.Context, input AppVersionUploadURLInput) (AppVersionUploadURLOutput, error) {
	var resp appVersionUploadURLResponse

	variables := map[string]any{
		"appVersionID": input.AppVersionID,
	}
	err := s.client.Exec(ctx, appVersionUploadURLText, &resp, variables)
	if err != nil {
		return AppVersionUploadURLOutput{}, err
	}

	return AppVersionUploadURLOutput{UploadURL: resp.AppVersionUploadURL.URL}, nil
}
