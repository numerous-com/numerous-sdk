package auth

// Credentials is used to facilitate the login process.
type Credentials struct {
	Audience            string
	ClientID            string
	DeviceCodeEndpoint  string
	OauthTokenEndpoint  string
	RevokeTokenEndpoint string
}
