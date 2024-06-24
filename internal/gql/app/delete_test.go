package app

import (
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	t.Run("can return status on AppDelete mutation", func(t *testing.T) {
		response, expectedType := test.DeleteSuccessQueryResult()
		c := test.CreateTestGqlClient(t, response)

		actualStatus, err := Delete("id", c)

		assert.NoError(t, err)
		assert.Equal(t, expectedType, actualStatus.ToolDelete.Typename)
	})

	t.Run("can return error on AppDelete mutation", func(t *testing.T) {
		appNotFoundResponse := "record not found"
		response, expectedType := test.DeleteFailureQueryResult(appNotFoundResponse)
		c := test.CreateTestGqlClient(t, response)
		actualStatus, _ := Delete("id", c)
		assert.Equal(t, expectedType, actualStatus.ToolDelete.Typename)
		assert.Equal(t, appNotFoundResponse, actualStatus.ToolDelete.Result)
	})
}
