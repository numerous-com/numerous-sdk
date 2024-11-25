package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type AppVersionUploadURLInput struct {
	AppVersionID string
}

type AppVersionUploadURLOutput struct {
	UploadURL string
}

const appVersionUploadURLText = `
mutation CLIAppVersionUploadURL($appVersionID: ID!) {
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
	err := s.client.Exec(ctx, appVersionUploadURLText, &resp, variables, graphql.OperationName("CLIAppVersionUploadURL"))
	if err != nil {
		return AppVersionUploadURLOutput{}, err
	}

	return AppVersionUploadURLOutput{UploadURL: resp.AppVersionUploadURL.URL}, nil
}
