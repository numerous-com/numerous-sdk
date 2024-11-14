package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
)

var ErrEmailNotVerified = errors.New("email not verified")

type Result struct {
	IDToken      string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// waitUntilUserLogsIn waits until the user is logged in on the browser.
func waitUntilUserLogsIn(ctx context.Context, httpClient *http.Client, t *time.Ticker, deviceCode string, cred Credentials) (Result, error) {
	for {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-t.C:
			data := url.Values{
				"client_id":   []string{cred.ClientID},
				"grant_type":  []string{"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code": []string{deviceCode},
			}
			r, err := httpClient.PostForm(cred.OauthTokenEndpoint, data)
			if err != nil {
				return Result{}, fmt.Errorf("cannot get device code: %w", err)
			}
			defer func() {
				_ = r.Body.Close()
			}()

			var res struct {
				AccessToken      string  `json:"access_token"`
				IDToken          string  `json:"id_token"`
				RefreshToken     string  `json:"refresh_token"`
				Scope            string  `json:"scope"`
				ExpiresIn        int64   `json:"expires_in"`
				TokenType        string  `json:"token_type"`
				Error            *string `json:"error,omitempty"`
				ErrorDescription string  `json:"error_description,omitempty"`
			}

			err = json.NewDecoder(r.Body).Decode(&res)
			if err != nil {
				return Result{}, fmt.Errorf("cannot decode response: %w", err)
			}

			if res.Error != nil {
				if *res.Error == "authorization_pending" {
					continue
				}
				if *res.Error == "access_denied" && strings.Contains(res.ErrorDescription, "email not verified.") {
					return Result{}, ErrEmailNotVerified
				}

				return Result{}, errors.New(res.ErrorDescription)
			}

			if err := validateAccessToken(res.AccessToken); err != nil {
				return Result{}, err
			}

			return Result{
				RefreshToken: res.RefreshToken,
				AccessToken:  res.AccessToken,
				IDToken:      res.IDToken,
				ExpiresAt: time.Now().Add(
					time.Duration(res.ExpiresIn) * time.Second,
				),
			}, nil
		}
	}
}

func validateAccessToken(accessToken string) error {
	_, err := jwt.ParseString(accessToken)
	if err != nil {
		return err
	}

	return nil
}
