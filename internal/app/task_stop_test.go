package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStopTask(t *testing.T) {
	t.Run("returns expected result", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"taskStop": {
						"taskInstanceID": "test-instance-id"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		taskInstanceID := "test-instance-id"
		result, err := s.StopTask(context.TODO(), taskInstanceID)

		expected := &TaskStopResult{
			TaskInstanceID: "test-instance-id",
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns error if mutation fails", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "Task instance not found",
					"location": [{"line": 1, "column": 1}],
					"path": ["taskStop"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		taskInstanceID := "nonexistent-instance"
		result, err := s.StopTask(context.TODO(), taskInstanceID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
