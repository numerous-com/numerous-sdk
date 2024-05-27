package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
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

func CreateVersion(ctx context.Context, client *graphql.Client, input CreateAppVersionInput) (CreateAppVersionOutput, error) {
	var resp appVersionCreateResponse

	variables := map[string]any{
		"appID": input.AppID,
	}
	err := client.Exec(ctx, appVersionCreateText, &resp, variables)
	if err != nil {
		return CreateAppVersionOutput{}, err
	}

	return CreateAppVersionOutput{
		AppVersionID: resp.AppVersionCreate.ID,
	}, nil
}
