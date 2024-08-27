package create

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/token"
)

func TestCreate(t *testing.T) {
	name := "token name"
	desc := "token description"
	testErr := errors.New("test error")

	t.Run("creates with expected name and description", func(t *testing.T) {
		mtc := MockTokenCreator{}
		mtc.On("Create", mock.Anything, token.CreateTokenInput{Name: name, Description: desc}).Return(token.CreateTokenOutput{}, nil)

		err := Create(context.TODO(), &mtc, CreateInput{Name: name, Description: desc})

		assert.NoError(t, err)
		mtc.AssertExpectations(t)
	})

	t.Run("passes on error", func(t *testing.T) {
		for _, expectedError := range []error{
			token.ErrAccessDenied,
			token.ErrUserAccessTokenAlreadyExists,
			token.ErrUserAccessTokenNameInvalid,
			testErr,
		} {
			t.Run(expectedError.Error(), func(t *testing.T) {
				mtc := MockTokenCreator{}
				mtc.On("Create", mock.Anything, token.CreateTokenInput{Name: name, Description: desc}).Return(token.CreateTokenOutput{}, expectedError)

				err := Create(context.TODO(), &mtc, CreateInput{Name: name, Description: desc})

				assert.ErrorIs(t, err, expectedError)
			})
		}
	})
}
