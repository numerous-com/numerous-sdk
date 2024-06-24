package secret

import "encoding/base64"

func AppSecretsFromMap(secrets map[string]string) []*AppSecret {
	convertedSecrets := make([]*AppSecret, 0)

	for name, value := range secrets {
		secret := &AppSecret{
			Name:        name,
			Base64Value: base64.StdEncoding.EncodeToString([]byte(value)),
		}
		convertedSecrets = append(convertedSecrets, secret)
	}

	return convertedSecrets
}

type AppSecret struct {
	Name        string `json:"name"`
	Base64Value string `json:"base64Value"`
}
