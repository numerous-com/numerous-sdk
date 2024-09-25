package token

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type RevokeTokenOutput struct {
	Name        string
	Description string
}

type personalAccessTokenRevokeResponse struct {
	PersonalAccessTokenRevoke struct {
		Name        string
		Description string
	} `graphql:"personalAccessTokenRevoke(id: $id)"`
}

func (s *Service) Revoke(ctx context.Context, id string) (RevokeTokenOutput, error) {
	var resp personalAccessTokenRevokeResponse

	if err := s.client.Mutate(ctx, &resp, map[string]interface{}{"id": graphql.ID(id)}); err != nil {
		return RevokeTokenOutput{}, ConvertErrors(err)
	} else {
		result := resp.PersonalAccessTokenRevoke
		return RevokeTokenOutput{
			Name:        result.Name,
			Description: result.Description,
		}, nil
	}
}
