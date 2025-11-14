package app

import (
	"encoding/base64"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeTaskDataForDisplay(t *testing.T) {
	t.Run("returns (none) for nil input", func(t *testing.T) {
		result := DecodeTaskDataForDisplay(nil)
		assert.Equal(t, "(none)", result)
	})

	t.Run("returns expected decoded string for valid base64 text", func(t *testing.T) {
		input := "Hello, World!"
		encoded := base64.StdEncoding.EncodeToString([]byte(input))
		result := DecodeTaskDataForDisplay(&encoded)
		assert.Equal(t, input, result)
	})

	t.Run("returns expected decoded string for valid base64 JSON", func(t *testing.T) {
		input := `{"user_id": 123, "action": "process"}`
		encoded := base64.StdEncoding.EncodeToString([]byte(input))
		result := DecodeTaskDataForDisplay(&encoded)
		assert.Equal(t, input, result)
	})

	t.Run("returns (base64) indicator for invalid base64", func(t *testing.T) {
		invalid := "not-valid-base64!"
		result := DecodeTaskDataForDisplay(&invalid)
		assert.Contains(t, result, "(base64)")
	})

	t.Run("returns (binary data) for non-UTF-8 content", func(t *testing.T) {
		binaryData := []byte{0xFF, 0xFE, 0xFD}
		encoded := base64.StdEncoding.EncodeToString(binaryData)
		result := DecodeTaskDataForDisplay(&encoded)
		assert.Equal(t, "(binary data)", result)
	})
}

func TestEncodeTaskInput(t *testing.T) {
	maxRawDataSize := 3 * int(math.Ceil(float64(MaxTaskInputSize/4)))
	t.Run("encodes simple string", func(t *testing.T) {
		input := "test input"
		encoded, err := encodeTaskInput(input)
		assert.NoError(t, err)

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		assert.NoError(t, err)
		assert.Equal(t, input, string(decoded))
	})

	t.Run("encodes JSON string", func(t *testing.T) {
		input := `{"user_id": 123}`
		encoded, err := encodeTaskInput(input)
		assert.NoError(t, err)

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		assert.NoError(t, err)
		assert.Equal(t, input, string(decoded))
	})

	t.Run("returns error when input exceeds max size", func(t *testing.T) {
		input := strings.Repeat("a", maxRawDataSize+1)
		_, err := encodeTaskInput(input)
		assert.Error(t, err)
		assert.Equal(t, ErrTaskInputTooLarge, err)
	})

	t.Run("returns expected encoded string when input is exactly max size", func(t *testing.T) {
		input := strings.Repeat("a", maxRawDataSize)
		encoded, err := encodeTaskInput(input)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(encoded), MaxTaskInputSize)
	})

	t.Run("handles empty string", func(t *testing.T) {
		input := ""
		encoded, err := encodeTaskInput(input)
		assert.NoError(t, err)
		assert.Equal(t, input, encoded)
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		input := "Hello, " + "\u2713"
		encoded, err := encodeTaskInput(input)
		assert.NoError(t, err)

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		assert.NoError(t, err)
		assert.Equal(t, input, string(decoded))
	})
}
