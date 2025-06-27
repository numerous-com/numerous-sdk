package login

import (
	"context"
	"testing"
	"time"

	"numerous.com/cli/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	testTenant := "test.domain.com"
	state := auth.DeviceCodeState{
		DeviceCode:      "some-code",
		UserCode:        "some-long-user-code",
		VerificationURI: "https://test.domain.com/device/code/some-code",
		ExpiresIn:       8400,
		Interval:        5,
	}
	result := auth.Result{
		IDToken:      "some-id-token",
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
		ExpiresAt:    time.Now().Add(time.Second * time.Duration(state.ExpiresIn)),
	}
	expectedUser := &auth.User{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Tenant:       testTenant,
	}

	m := new(auth.MockAuthenticator)
	m.On("GetDeviceCode", mock.Anything, mock.Anything).Return(state, nil)
	m.On("OpenURL", state.VerificationURI).Return(nil)
	m.On("WaitUntilUserLogsIn", mock.Anything, mock.Anything, state).Return(result, nil)
	m.On("StoreBothTokens", result.AccessToken, result.RefreshToken).Return(nil)
	m.On("GetLoggedInUserFromKeyring").Return(&auth.User{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Tenant:       testTenant,
	})
	acutalUser, _ := login(m, context.Background())

	m.AssertExpectations(t)
	assert.Equal(t, expectedUser, acutalUser)
}
