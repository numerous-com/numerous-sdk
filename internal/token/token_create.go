package token

import (
	"context"
	"fmt"
)

type CreateTokenInput struct {
	Name        string
	Description string
}

type CreateTokenOutput struct {
	Name        string
	Description string
	Token       string
}

type userAccessTokenCreateResponse struct {
	UserAccessTokenCreate struct {
		Typename string `graphql:"__typename"`
		Created  struct {
			Entry struct {
				Name        string
				Description string
			}
			Token string
		} `graphql:"... on UserAccessTokenCreated"`
		InvalidName struct {
			Name   string
			Reason string
		} `graphql:"... on UserAccessTokenInvalidName"`
		AlreadyExists struct{ Name string } `graphql:"... on UserAccessTokenAlreadyExists"`
	} `graphql:"userAccessTokenCreate(input: {name: $name, description: $desc})"`
}

func (s *Service) Create(ctx context.Context, input CreateTokenInput) (CreateTokenOutput, error) {
	var resp userAccessTokenCreateResponse

	err := s.client.Mutate(ctx, &resp, map[string]interface{}{"name": input.Name, "desc": input.Description})
	if err != nil {
		return CreateTokenOutput{}, ConvertErrors(err)
	}

	result := resp.UserAccessTokenCreate
	switch result.Typename {
	case "UserAccessTokenCreated":
		return CreateTokenOutput{
			Name:        result.Created.Entry.Name,
			Description: result.Created.Entry.Description,
			Token:       result.Created.Token,
		}, nil
	case "UserAccessTokenInvalidName":
		return CreateTokenOutput{}, fmt.Errorf("%w: %s", ErrUserAccessTokenNameInvalid, result.InvalidName.Reason)
	case "UserAccessTokenAlreadyExists":
		return CreateTokenOutput{}, fmt.Errorf("%w: %s", ErrUserAccessTokenAlreadyExists, result.AlreadyExists.Name)
	default:
		panic("unexpected response from server")
	}
}
