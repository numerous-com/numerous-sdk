package gql

import (
	"net/http"
	"os"
	"sync"

	"numerous.com/cli/internal/auth"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/hasura/go-graphql-client"
)

var (
	client *gqlclient.Client
	once   sync.Once
)

var _ http.RoundTripper = &AuthenticatingRoundTripper{}

type AuthenticatingRoundTripper struct {
	proxied     http.RoundTripper
	accessToken *string
}

// RoundTrip implements http.RoundTripper.
func (a *AuthenticatingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if a.accessToken != nil {
		r.Header.Set("Authorization", "Bearer "+*a.accessToken)
	}

	return a.proxied.RoundTrip(r)
}

func initClient() {
	httpClient := &http.Client{
		Transport: &AuthenticatingRoundTripper{
			proxied:     http.DefaultTransport,
			accessToken: getAccessToken(),
		},
	}

	client = gqlclient.New(getHTTPURL(), httpClient)
}

func GetClient() *gqlclient.Client {
	once.Do(initClient)
	return client
}

func NewClient() *graphql.Client {
	client := graphql.NewClient(getHTTPURL(), http.DefaultClient)

	accessToken := getAccessToken()
	if accessToken != nil {
		client = client.WithRequestModifier(func(r *http.Request) {
			if accessToken != nil {
				r.Header.Set("Authorization", "Bearer "+*accessToken)
			}
		})
	}

	return client
}

func getAccessToken() *string {
	token := os.Getenv("NUMEROUS_ACCESS_TOKEN")
	if token != "" {
		return &token
	}

	user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
	if user != nil {
		return &user.AccessToken
	}

	return nil
}
