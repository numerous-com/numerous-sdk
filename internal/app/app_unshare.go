package app

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"numerous.com/cli/internal/appident"
)

type appDeployUnshareResponse struct {
	AppDeployUnshare struct {
		Typename string `graphql:"__typename"`
	} `graphql:"appDeployUnshare(deployID: $deployID)"`
}

func (s *Service) UnshareApp(ctx context.Context, ai appident.AppIdentifier) error {
	var appResp queryAppDefaultDeployID
	if err := s.client.Query(ctx, &appResp, map[string]interface{}{"orgSlug": ai.OrganizationSlug, "appSlug": ai.AppSlug}); err != nil {
		return convertErrors(err)
	}

	var appUnshareResp appDeployUnshareResponse
	if err := s.client.Mutate(ctx, &appUnshareResp, map[string]interface{}{"deployID": graphql.ID(appResp.App.DefaultDeployment.ID)}); err != nil {
		return convertErrors(err)
	} else {
		return nil
	}
}
