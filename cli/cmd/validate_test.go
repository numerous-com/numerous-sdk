package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIdent(t *testing.T) {
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
			actual := validIdent(invalid)
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
			actual := validIdent(valid)
			assert.True(t, actual)
		})
	}
}
