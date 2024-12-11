package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

func TestJobsByTool(t *testing.T) {
	t.Run("returns expected jobs", func(t *testing.T) {
		client := test.CreateTestGqlClient(t, `{"data": {"jobsByTool": [{"id": "job-id-1"}, {"id": "job-id-2"}]}}`)

		actual, err := JobsByTool("tool-id", client)

		assert.NoError(t, err)
		assert.Equal(t, []Job{{ID: "job-id-1"}, {ID: "job-id-2"}}, actual)
	})
}
