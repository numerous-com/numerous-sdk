package list

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/token"
)

func TestList(t *testing.T) {
	testErr := errors.New("test error")

	t.Run("lists with no errors on expected response", func(t *testing.T) {
		expirationTime, err := time.Parse(time.RFC3339, "2026-09-27T11:12:13.123456Z")
		require.NoError(t, err)

		out := token.ListTokenOutput{
			{
				ID:          "first-token-id",
				Name:        "First token name",
				Description: "first token description",
				ExpiresAt:   &expirationTime,
			},
			{
				ID:          "second-token-id",
				Name:        "Second token name",
				Description: "second token description",
				ExpiresAt:   nil,
			},
		}
		lister := MockTokenLister{}
		lister.On("List", mock.Anything).Return(out, nil)

		err = List(context.TODO(), &lister)

		assert.NoError(t, err)
		lister.AssertExpectations(t)
	})

	t.Run("passes on error", func(t *testing.T) {
		for _, expectedError := range []error{
			token.ErrAccessDenied,
			testErr,
		} {
			t.Run(expectedError.Error(), func(t *testing.T) {
				lister := MockTokenLister{}
				lister.On("List", mock.Anything).Return(token.ListTokenOutput{}, expectedError)

				err := List(context.TODO(), &lister)

				assert.ErrorIs(t, err, expectedError)
			})
		}
	})
}
