package gql

import (
	"net/http"
	"net/url"
	"sync"

	"numerous/cli/auth"

	"git.sr.ht/~emersion/gqlclient"
)

var (
	client *gqlclient.Client
	once   sync.Once
)

func initClient() {
	var httpClient *http.Client

	if user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring(); user != nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*url.URL, error) {
					r.Header.Set("Authorization", "Bearer "+user.AccessToken)
					return nil, nil
				},
			},
		}
	} else {
		httpClient = http.DefaultClient
	}

	client = gqlclient.New(httpURL, httpClient)
}

func GetClient() *gqlclient.Client {
	once.Do(initClient)
	return client
}
