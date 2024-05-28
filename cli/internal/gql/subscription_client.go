package gql

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

func GetSubscriptionClient() *graphql.SubscriptionClient {
	client := graphql.NewSubscriptionClient(wsURL)

	accessToken := getAccessToken()
	if accessToken != nil {
		client = client.WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPHeader: http.Header{
				"Authorization": []string{"Bearer " + *accessToken},
			},
		})
	}

	return client
}
