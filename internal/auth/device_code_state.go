package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const waitThresholdInSeconds = 3

type DeviceCodeState struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri_complete"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (s *DeviceCodeState) IntervalDuration() time.Duration {
	return time.Duration(s.Interval+waitThresholdInSeconds) * time.Second
}

var scope = []string{"openid", "profile", "offline_access", "email"}

func getDeviceCodeState(ctx context.Context, httpClient *http.Client, c Credentials) (DeviceCodeState, error) {
	data := url.Values{
		"client_id": []string{c.ClientID},
		"scope":     []string{strings.Join(scope, " ")},
		"audience":  []string{c.Audience},
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.DeviceCodeEndpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return DeviceCodeState{}, fmt.Errorf("failed to create the request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpClient.Do(request)
	if err != nil {
		return DeviceCodeState{}, fmt.Errorf("failed to send the request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return DeviceCodeState{}, fmt.Errorf(
				"received a %d response and failed to read the response",
				response.StatusCode,
			)
		}

		return DeviceCodeState{}, fmt.Errorf("received a %d response: %s", response.StatusCode, bodyBytes)
	}

	var state DeviceCodeState
	if err = json.NewDecoder(response.Body).Decode(&state); err != nil {
		return DeviceCodeState{}, fmt.Errorf("failed to decode the response: %w", err)
	}

	return state, nil
}
