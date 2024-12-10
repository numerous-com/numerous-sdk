package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"numerous.com/cli/internal/keyring"

	"github.com/lestrrat-go/jwx/jwt"
)

var (
	// ErrUserNotLoggedIn is thrown when there is not Access or Refresh Token
	ErrUserNotLoggedIn = errors.New("user is logged in")
	ErrInvalidToken    = errors.New("token is invalid")
	ErrExpiredToken    = errors.New("token is expired")
)

type User struct {
	AccessToken  string
	RefreshToken string
	Tenant       string
}

func getLoggedInUserFromKeyring(tenant string) *User {
	a, accessTokenErr := keyring.GetAccessToken(tenant)
	r, refreshTokenErr := keyring.GetRefreshToken(tenant)
	if accessTokenErr != nil || refreshTokenErr != nil {
		return nil
	}

	return &User{
		AccessToken:  a,
		RefreshToken: r,
		Tenant:       tenant,
	}
}

func (u *User) extractToken() (jwt.Token, error) {
	token, err := jwt.ParseString(u.AccessToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (u *User) CheckAuthenticationStatus() error {
	if u == nil || u.AccessToken == "" {
		return ErrUserNotLoggedIn
	}

	if token, err := u.extractToken(); err != nil {
		return ErrInvalidToken
	} else if err := validateToken(token, u.Tenant); err != nil {
		return err
	}

	return nil
}

func (u *User) HasExpiredToken() bool {
	token, _ := u.extractToken()
	return tokenExpired(token)
}

func (u *User) RefreshAccessToken(client *http.Client, a Authenticator) error {
	if err := u.CheckAuthenticationStatus(); err != ErrExpiredToken {
		if err != nil {
			return err
		}

		return nil
	}

	newAccessToken, err := a.RegenerateAccessToken(client, u.RefreshToken)
	if err != nil {
		return err
	}
	if err := a.StoreAccessToken(newAccessToken); err != nil {
		return err
	}
	u.AccessToken = newAccessToken

	return nil
}

func validateToken(t jwt.Token, tenantName string) error {
	err := jwt.Validate(t, jwt.WithIssuer(fmt.Sprintf("https://%s/", tenantName)))
	switch err {
	case jwt.ErrTokenExpired():
		return ErrExpiredToken
	case jwt.ErrInvalidIssuedAt():
		return ErrInvalidToken
	case nil:
		return nil
	default:
		return err
	}
}

func tokenExpired(t jwt.Token) bool {
	if t == nil {
		return false
	}

	return time.Now().After(t.Expiration())
}
