package app

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type UploadDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type SubscriptionClient interface {
	Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error)
	Run() error
	Close() error
}

type Service struct {
	client       *graphql.Client
	subscription SubscriptionClient
	uploadDoer   UploadDoer
}

func New(client *graphql.Client, subscription SubscriptionClient, uploadDoer UploadDoer) *Service {
	return &Service{
		client:       client,
		subscription: subscription,
		uploadDoer:   uploadDoer,
	}
}
