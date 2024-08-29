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

type personalAccessTokenCreateResponse struct {
	PersonalAccessTokenCreate struct {
		Typename string `graphql:"__typename"`
		Created  struct {
			Entry struct {
				Name        string
				Description string
			}
			Token string
		} `graphql:"... on PersonalAccessTokenCreated"`
		InvalidName struct {
			Name   string
			Reason string
		} `graphql:"... on PersonalAccessTokenInvalidName"`
		AlreadyExists struct{ Name string } `graphql:"... on PersonalAccessTokenAlreadyExists"`
	} `graphql:"personalAccessTokenCreate(input: {name: $name, description: $desc})"`
}

func (s *Service) Create(ctx context.Context, input CreateTokenInput) (CreateTokenOutput, error) {
	var resp personalAccessTokenCreateResponse

	err := s.client.Mutate(ctx, &resp, map[string]interface{}{"name": input.Name, "desc": input.Description})
	if err != nil {
		return CreateTokenOutput{}, ConvertErrors(err)
	}

	result := resp.PersonalAccessTokenCreate
	switch result.Typename {
	case "PersonalAccessTokenCreated":
		return CreateTokenOutput{
			Name:        result.Created.Entry.Name,
			Description: result.Created.Entry.Description,
			Token:       result.Created.Token,
		}, nil
	case "PersonalAccessTokenInvalidName":
		return CreateTokenOutput{}, fmt.Errorf("%w: %s", ErrPersonalAccessTokenNameInvalid, result.InvalidName.Reason)
	case "PersonalAccessTokenAlreadyExists":
		return CreateTokenOutput{}, fmt.Errorf("%w: %s", ErrPersonalAccessTokenAlreadyExists, result.AlreadyExists.Name)
	default:
		panic("unexpected response from server")
	}
}
