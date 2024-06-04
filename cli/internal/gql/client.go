package gql

import (
	"net/http"
	"sync"

	"numerous/cli/auth"

	"git.sr.ht/~emersion/gqlclient"
)

var (
	client *gqlclient.Client
	once   sync.Once
)

var _ http.RoundTripper = &AuthenticatingRoundTripper{}

type AuthenticatingRoundTripper struct {
	proxied http.RoundTripper
	user    *auth.User
}

// RoundTrip implements http.RoundTripper.
func (a *AuthenticatingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if a.user != nil {
		r.Header.Set("Authorization", "Bearer "+a.user.AccessToken)
	}

	return a.proxied.RoundTrip(r)
}

func initClient() {
	var httpClient *http.Client

	user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
	httpClient = &http.Client{
		Transport: &AuthenticatingRoundTripper{
			proxied: http.DefaultTransport,
			user:    user,
		},
	}

	client = gqlclient.New(httpURL, httpClient)
}

func GetClient() *gqlclient.Client {
	once.Do(initClient)
	return client
}
