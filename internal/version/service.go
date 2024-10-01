package version

import "github.com/hasura/go-graphql-client"

type Service struct {
	client *graphql.Client
}

func NewService(client *graphql.Client) *Service {
	return &Service{client: client}
}
