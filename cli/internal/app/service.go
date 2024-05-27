package app

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type UploadDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Service struct {
	client     *graphql.Client
	uploadDoer UploadDoer
}

func New(client *graphql.Client, uploadDoer UploadDoer) *Service {
	return &Service{
		client:     client,
		uploadDoer: uploadDoer,
	}
}
