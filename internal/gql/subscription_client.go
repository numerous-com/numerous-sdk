package gql

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

func NewSubscriptionClient() *graphql.SubscriptionClient {
	client := graphql.NewSubscriptionClient(GetWSURL())

	accessToken := GetAccessToken()
	if accessToken != nil {
		client = client.WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPHeader: http.Header{
				"Authorization": []string{"Bearer " + *accessToken},
			},
		})
	}

	return client
}
