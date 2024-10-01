package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/test"
)

func TestList(t *testing.T) {
	t.Run("returns expected apps", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
		{
			"data": {
				"organization": {
					"__typename": "Organization",
					"apps": [
						{
							"displayName": "App 1",
							"slug": "app-1",
							"description": "App 1 description",
							"createdBy": {"fullName": "User Name"},
							"createdAt": "2024-06-12T14:16:17.000Z",
							"defaultDeployment": {
								"current": {
									"status": "RUNNING"
								}
							}
						},
						{
							"displayName": "App 2",
							"slug": "app-2",
							"description": "App 2 description",
							"createdBy": {"fullName": "User Name"},
							"createdAt": "2024-05-10T12:14:16.000Z",
							"defaultDeployment": null
						},
						{
							"displayName": "App 3",
							"slug": "app-3",
							"description": "App 3 description",
							"createdBy": {"fullName": "User Name"},
							"createdAt": "2024-05-11T13:15:17.000Z",
							"defaultDeployment": {
								"current": null
							}
						}
					]
				}
			}
		}
	`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		apps, err := s.List(context.TODO(), "organization-slug")

		expected := []ListApp{
			{Name: "App 1", Slug: "app-1", Description: "App 1 description", Status: "RUNNING", CreatedBy: "User Name", CreatedAt: time.Date(2024, time.June, 12, 14, 16, 17, 0, time.UTC)},
			{Name: "App 2", Slug: "app-2", Description: "App 2 description", Status: "NOT DEPLOYED", CreatedBy: "User Name", CreatedAt: time.Date(2024, time.May, 10, 12, 14, 16, 0, time.UTC)},
			{Name: "App 3", Slug: "app-3", Description: "App 3 description", Status: "NOT DEPLOYED", CreatedBy: "User Name", CreatedAt: time.Date(2024, time.May, 11, 13, 15, 17, 0, time.UTC)},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, apps)
	})

	t.Run("returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)
		respBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["organization"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		apps, err := s.List(context.TODO(), "organization-slug")

		assert.Nil(t, apps)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("returns organization not found error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)
		respBody := `
			{
				"data": {
					"organization": {
						"__typename": "OrganizationNotFound"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		apps, err := s.List(context.TODO(), "organization-slug")

		assert.Nil(t, apps)
		assert.ErrorIs(t, err, ErrOrganizationNotFound)
	})

	t.Run("returns unexpected type error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)
		respBody := `
			{
				"data": {
					"organization": {
						"__typename": "SomeUnexpectedType"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		apps, err := s.List(context.TODO(), "organization-slug")

		assert.Nil(t, apps)
		assert.ErrorIs(t, err, ErrUnexpectedType)
	})
}
