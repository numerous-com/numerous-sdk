package token

import "github.com/hasura/go-graphql-client"

type Service struct {
	client *graphql.Client
}

func New(client *graphql.Client) *Service {
	return &Service{client: client}
}
