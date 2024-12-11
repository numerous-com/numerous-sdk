package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

func TestJobStop(t *testing.T) {
	t.Run("returns message from response", func(t *testing.T) {
		client := test.CreateTestGqlClient(t, `{"data": {"message": "test message"}}`)

		msg, err := JobStop("job-id", client)

		assert.NoError(t, err)
		assert.Equal(t, "test message", msg)
	})
}
