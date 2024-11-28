package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"numerous.com/cli/internal/appident"
)

type ShareAppOutput struct {
	SharedURL *string
}

type appDeployShareResponse struct {
	AppDeployShare struct {
		SharedURL *string `graphql:"sharedURL"`
	} `graphql:"appDeployShare(deployID: $deployID)"`
}

type queryAppDefaultDeployID struct {
	App struct {
		DefaultDeployment struct {
			ID string
		}
	} `graphql:"app(organizationSlug: $orgSlug, appSlug: $appSlug)"`
}

func (s *Service) ShareApp(ctx context.Context, ai appident.AppIdentifier) (ShareAppOutput, error) {
	var appResp queryAppDefaultDeployID
	if err := s.client.Query(ctx, &appResp, map[string]interface{}{"orgSlug": ai.OrganizationSlug, "appSlug": ai.AppSlug}); err != nil {
		return ShareAppOutput{}, convertErrors(err)
	}

	var appShareResp appDeployShareResponse
	if err := s.client.Mutate(ctx, &appShareResp, map[string]interface{}{"deployID": graphql.ID(appResp.App.DefaultDeployment.ID)}); err != nil {
		return ShareAppOutput{}, convertErrors(err)
	} else {
		return ShareAppOutput{SharedURL: appShareResp.AppDeployShare.SharedURL}, nil
	}
}
