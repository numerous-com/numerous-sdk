package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidIdentifier(t *testing.T) {
	for _, invalid := range []string{
		"abcæøå",
		"abc[123",
		"abc]123",
		"abc_123",
		"abc|123",
		"abc\"123",
		"abc'123",
		"abc<123",
		"abc>123",
		"abc#123",
		"abc?123",
		"abc:123",
		"abc;123",
		"abcABC123",
	} {
		t.Run("fails for '"+""+invalid+"'", func(t *testing.T) {
			actual := IsValidIdentifier(invalid)
			assert.False(t, actual)
		})
	}

	for _, valid := range []string{
		"abcdefghijklmopqrstuvwxyz",
		"1234567890",
		"abcdef123456",
		"abcdef-123456",
		"123456abcdef",
		"123456-abcdef",
	} {
		t.Run("succeeds for '"+valid+"'", func(t *testing.T) {
			actual := IsValidIdentifier(valid)
			assert.True(t, actual)
		})
	}
}
