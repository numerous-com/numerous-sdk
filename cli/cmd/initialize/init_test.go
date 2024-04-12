package initialize

import (
	"testing"

	"numerous/cli/tool"

	"github.com/stretchr/testify/assert"
)

func TestValidateFlags(t *testing.T) {
	t.Run("Validates if no library is set", func(t *testing.T) {
		err := validateAndSetAppLibrary(&tool.Tool{}, "")
		assert.NoError(t, err)
	})

	t.Run("Cannot validate unsupported library", func(t *testing.T) {
		err := validateAndSetAppLibrary(&tool.Tool{}, "something")
		assert.Error(t, err)
	})

	for _, lib := range []string{"plotly", "marimo", "streamlit"} {
		t.Run("Validates "+lib, func(t *testing.T) {
			err := validateAndSetAppLibrary(&tool.Tool{}, lib)
			assert.NoError(t, err)
		})
	}
}
