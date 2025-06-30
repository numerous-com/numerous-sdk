package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type CreateAppVersionInput struct {
	AppID   string
	Version string
	Message string
	Size    *string
}

type CreateAppVersionOutput struct {
	AppVersionID     string
	OrganizationSlug string
	AppSlug          string
}

const appVersionCreateText = `
mutation CLIAppVersionCreate($appID: ID!, $version: String, $message: String!, $size: String) {
	appVersionCreate(appID: $appID, input: {version: $version, message: $message, size: $size}) {
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
		"size":    input.Size,
	}

	if input.Version != "" {
		variables["version"] = input.Version
	}

	err := s.client.Exec(ctx, appVersionCreateText, &resp, variables, graphql.OperationName("CLIAppVersionCreate"))
	if err != nil {
		return CreateAppVersionOutput{}, err
	}

	return CreateAppVersionOutput{AppVersionID: resp.AppVersionCreate.ID}, nil
}
