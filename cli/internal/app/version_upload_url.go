package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type GetAppVersionUploadURLInput struct {
	AppVersionID string
}

type GetAppVersionUploadURLOutput struct {
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

func GetAppVersionUploadURL(ctx context.Context, client *graphql.Client, input GetAppVersionUploadURLInput) (GetAppVersionUploadURLOutput, error) {
	var resp appVersionUploadURLResponse

	variables := map[string]any{
		"appVersionID": input.AppVersionID,
	}
	err := client.Exec(ctx, appVersionUploadURLText, &resp, variables)
	if err != nil {
		return GetAppVersionUploadURLOutput{}, err
	}

	return GetAppVersionUploadURLOutput{UploadURL: resp.AppVersionUploadURL.URL}, nil
}
