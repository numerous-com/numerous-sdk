package gql

import (
	"net/http"

	"numerous/cli/auth"

	"github.com/hasura/go-graphql-client"
)

func GetSubscriptionClient() *graphql.SubscriptionClient {
	client := graphql.NewSubscriptionClient(wsURL)

	if user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring(); user != nil {
		client = client.WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPHeader: http.Header{
				"Authorization": []string{"Bearer " + user.AccessToken},
			},
		})
	}

	return client
}
