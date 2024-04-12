package mocks

import (
	"context"
	"net/http"

	"numerous/cli/auth"

	"github.com/stretchr/testify/mock"
)

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) GetDeviceCode(ctx context.Context, client *http.Client) (auth.DeviceCodeState, error) {
	args := m.Called(ctx, client)
	return args.Get(0).(auth.DeviceCodeState), args.Error(1)
}

func (m *MockAuthenticator) OpenURL(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockAuthenticator) WaitUntilUserLogsIn(ctx context.Context, client *http.Client, state auth.DeviceCodeState) (auth.Result, error) {
	args := m.Called(ctx, client, state)
	return args.Get(0).(auth.Result), args.Error(1)
}

func (m *MockAuthenticator) StoreAccessToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthenticator) StoreRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthenticator) GetLoggedInUserFromKeyring() *auth.User {
	args := m.Called()
	return args.Get(0).(*auth.User)
}

func (m *MockAuthenticator) RegenerateAccessToken(client *http.Client, refreshToken string) (string, error) {
	args := m.Called(client, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthenticator) RevokeRefreshToken(client *http.Client, refreshToken string) error {
	args := m.Called(client, refreshToken)
	return args.Error(0)
}

func (m *MockAuthenticator) RemoveLoggedInUserFromKeyring() error {
	args := m.Called()
	return args.Error(0)
}
