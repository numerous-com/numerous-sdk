package revoke

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/token"
)

func TestRevoke(t *testing.T) {
	id := "test-token-id"
	name := "token name"
	description := "token description"
	testErr := errors.New("test error")

	t.Run("revokes and returns expected name and description", func(t *testing.T) {
		revoker := MockTokenRevoker{}
		revoker.On("Revoke", mock.Anything, id).Return(token.RevokeTokenOutput{Name: name, Description: description}, nil)

		err := Revoke(context.TODO(), &revoker, id)

		assert.NoError(t, err)
		revoker.AssertExpectations(t)
	})

	t.Run("passes on error", func(t *testing.T) {
		for _, expectedError := range []error{
			token.ErrAccessDenied,
			testErr,
		} {
			t.Run(expectedError.Error(), func(t *testing.T) {
				revoker := MockTokenRevoker{}
				revoker.On("Revoke", mock.Anything, id).Return(token.RevokeTokenOutput{}, expectedError)

				err := Revoke(context.TODO(), &revoker, id)

				assert.ErrorIs(t, err, expectedError)
			})
		}
	})
}
