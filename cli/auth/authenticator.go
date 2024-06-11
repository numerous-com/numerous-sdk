package auth

import (
	"context"
	"net/http"
	"time"

	"numerous/cli/keyring"

	"github.com/pkg/browser"
)

var NumerousTenantAuthenticator = NewTenantAuthenticator(auth0Domain, auth0ClientID)

type Authenticator interface {
	GetDeviceCode(ctx context.Context, client *http.Client) (DeviceCodeState, error)
	OpenURL(url string) error
	WaitUntilUserLogsIn(ctx context.Context, client *http.Client, state DeviceCodeState) (Result, error)
	StoreAccessToken(token string) error
	StoreRefreshToken(token string) error
	GetLoggedInUserFromKeyring() *User
	RemoveLoggedInUserFromKeyring() error
	RegenerateAccessToken(client *http.Client, refreshToken string) (string, error)
	RevokeRefreshToken(client *http.Client, refreshToken string) error
}

type TenantAuthenticator struct {
	tenant      string
	credentials Credentials
}

func NewTenantAuthenticator(tenant string, clientID string) *TenantAuthenticator {
	baseURL := "https://" + tenant
	return &TenantAuthenticator{
		tenant: tenant,
		credentials: Credentials{
			ClientID:            clientID,
			Audience:            baseURL + "/api/v2/",
			DeviceCodeEndpoint:  baseURL + "/oauth/device/code/",
			OauthTokenEndpoint:  baseURL + "/oauth/token/",
			RevokeTokenEndpoint: baseURL + "/oauth/revoke/",
		},
	}
}

func (t *TenantAuthenticator) GetDeviceCode(ctx context.Context, client *http.Client) (DeviceCodeState, error) {
	return getDeviceCodeState(ctx, client, t.credentials)
}

func (*TenantAuthenticator) OpenURL(url string) error {
	return browser.OpenURL(url)
}

func (t *TenantAuthenticator) WaitUntilUserLogsIn(ctx context.Context, client *http.Client, state DeviceCodeState) (Result, error) {
	ticker := time.NewTicker(state.IntervalDuration())
	return waitUntilUserLogsIn(ctx, client, ticker, state.DeviceCode, t.credentials)
}

func (t *TenantAuthenticator) StoreAccessToken(token string) error {
	return keyring.StoreAccessToken(t.tenant, token)
}

func (t *TenantAuthenticator) StoreRefreshToken(token string) error {
	return keyring.StoreRefreshToken(t.tenant, token)
}

func (t *TenantAuthenticator) GetLoggedInUserFromKeyring() *User {
	return getLoggedInUserFromKeyring(t.tenant)
}

func (t *TenantAuthenticator) RegenerateAccessToken(client *http.Client, refreshToken string) (string, error) {
	token, err := refreshAccessToken(client, refreshToken, t.credentials)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func (t *TenantAuthenticator) RevokeRefreshToken(client *http.Client, refreshToken string) error {
	return revokeRefreshToken(client, refreshToken, t.credentials)
}

func (t *TenantAuthenticator) RemoveLoggedInUserFromKeyring() error {
	return keyring.DeleteTokens(t.tenant)
}
