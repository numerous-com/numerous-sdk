package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type AppVersionDownloadURLInput struct {
	AppVersionID string
}

type AppVersionDownloadURLOutput struct {
	DownloadURL string
}

const appVersionDownloadURLText = `
mutation CLIAppVersionDownloadURL($appVersionID: ID!) {
	appVersionDownloadURL(appVersionID: $appVersionID) {
		url
	}
}
`

type appVersionDownloadURLResponse struct {
	AppVersionDownloadURL struct {
		URL string
	}
}

func (s *Service) AppVersionDownloadURL(ctx context.Context, input AppVersionDownloadURLInput) (AppVersionDownloadURLOutput, error) {
	var resp appVersionDownloadURLResponse

	variables := map[string]any{
		"appVersionID": input.AppVersionID,
	}
	err := s.client.Exec(ctx, appVersionDownloadURLText, &resp, variables, graphql.OperationName("CLIAppVersionDownloadURL"))
	if err != nil {
		return AppVersionDownloadURLOutput{}, convertErrors(err)
	}

	return AppVersionDownloadURLOutput{DownloadURL: resp.AppVersionDownloadURL.URL}, nil
}
