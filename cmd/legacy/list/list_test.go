package list

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	testUser := &auth.User{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Tenant:       "numerous-testing.com",
	}

	t.Run("Can list apps if user is signed in", func(t *testing.T) {
		m := new(auth.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(testUser)

		appsAsStrings := []string{}
		for i := 0; i < 3; i++ {
			id := strconv.Itoa(i)
			app := app.App{
				ID:        id,
				SharedURL: "https://test.com/shared/some-hash-" + id,
				PublicURL: "https://test.com/public/other-hash-" + id,
				Name:      "Name " + id,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			appsAsStrings = append(appsAsStrings, test.AppToResponse(app))
		}

		response := `{
			"data": {
				"tools": [` + strings.Join(appsAsStrings, ", ") + `]
			}
		}`
		c, transportMock := test.CreateMockGqlClient(response)

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})

	t.Run("Does not query list if user is not signed in", func(t *testing.T) {
		var u *auth.User
		m := new(auth.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		c, transportMock := test.CreateMockGqlClient()

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})

	t.Run("returns empty if url is empty", func(t *testing.T) {
		assert.Equal(t, "", getPublicEmoji(""))
	})

	t.Run("returns checkmark if url is not empty", func(t *testing.T) {
		assert.Equal(t, "✅", getPublicEmoji("https://test.com/public/id"))
	})
}
