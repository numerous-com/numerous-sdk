package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidClient  = errors.New("invalid client")
	ErrUnexpected     = errors.New("unexpected error")
)

type revokeResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func revokeRefreshToken(httpClient *http.Client, refreshToken string, cred Credentials) error {
	r, err := httpClient.PostForm(cred.RevokeTokenEndpoint, url.Values{
		"client_id": {cred.ClientID},
		"token":     {refreshToken},
	})
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		return nil
	}

	res := revokeResponse{}
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return err
	}

	switch res.Error {
	case "invalid_request":
		return ErrInvalidRequest
	case "invalid_client":
		return ErrInvalidClient
	default:
		return ErrUnexpected
	}
}
