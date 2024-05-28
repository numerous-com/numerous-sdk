package app

import (
	"context"
	"errors"
)

type ReadAppInput struct {
	OrganizationSlug string
	Name             string
}

type ReadAppOutput struct {
	AppID string
}

const queryAppText = `
query App($slug: String!, $name: String!) {
	app(organizationSlug: $slug, appName: $name) {
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

	variables := map[string]any{"slug": input.OrganizationSlug, "name": input.Name}
	err := s.client.Exec(ctx, queryAppText, &resp, variables)
	if err != nil {
		return ReadAppOutput{}, err
	}

	empty := appResponse{}
	if resp == empty {
		return ReadAppOutput{}, ErrAppNotFound
	}

	return ReadAppOutput{AppID: resp.App.ID}, nil
}
