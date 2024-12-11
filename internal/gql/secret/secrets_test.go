package secret

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppSecretsFromMap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		actual := AppSecretsFromMap(map[string]string{
			"SECRET_1": "value 1",
			"SECRET_2": "value 2",
		})

		assert.Contains(t, actual, &AppSecret{Name: "SECRET_1", Base64Value: base64.StdEncoding.EncodeToString([]byte("value 1"))})
		assert.Contains(t, actual, &AppSecret{Name: "SECRET_2", Base64Value: base64.StdEncoding.EncodeToString([]byte("value 2"))})
		assert.Len(t, actual, 2)
	})
}
