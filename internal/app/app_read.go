package app

import (
	"context"
	"errors"
	"strings"

	"numerous.com/cli/internal/gql"
)

type ReadAppInput struct {
	OrganizationSlug string
	AppSlug          string
}

type ReadAppOutput struct {
	AppID string
}

const queryAppText = `
query App($orgSlug: String!, $appSlug: String!) {
	app(organizationSlug: $orgSlug, appSlug: $appSlug) {
		id
	}
}
`

type appResponse struct {
	App struct {
		ID string
	}
}

var ErrAppNotFound = errors.New("app not found")

func (s *Service) ReadApp(ctx context.Context, input ReadAppInput) (ReadAppOutput, error) {
	var resp appResponse

	variables := map[string]any{"orgSlug": input.OrganizationSlug, "appSlug": input.AppSlug}
	err := s.client.Exec(ctx, queryAppText, &resp, variables)
	if err == nil {
		return ReadAppOutput{AppID: resp.App.ID}, nil
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "app not found") {
		return ReadAppOutput{}, ErrAppNotFound
	}

	return ReadAppOutput{}, gql.CheckAccessDenied(err)
}
