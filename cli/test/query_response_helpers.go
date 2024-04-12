package test

import (
	"fmt"
	"time"
)

func AppToQueryResult(queryName string, a struct {
	ID          string
	Name        string
	Description string
	PublicURL   string
	SharedURL   string
	CreatedAt   time.Time
},
) string {
	return fmt.Sprintf(`{
		"data": {
			"%s": %s
		}
	}`, queryName, AppToResponse(a))
}

func AppToResponse(a struct {
	ID          string
	Name        string
	Description string
	PublicURL   string
	SharedURL   string
	CreatedAt   time.Time
},
) string {
	description := quoteStringOrNull(a.Description)
	sharedURL := quoteStringOrNull(a.SharedURL)
	publicURL := quoteStringOrNull(a.PublicURL)
	createdAt := a.CreatedAt.Format(time.RFC3339)

	return fmt.Sprintf(`{
		"id": "%s",
		"name": "%s",
		"description": %s,
		"createdAt": "%s",
		"publicUrl": %s,
		"sharedUrl": %s
	}`, a.ID, a.Name, description, createdAt, publicURL, sharedURL)
}

type Role string

const (
	Admin Role = "ADMIN"
	User  Role = "USER"
)

func OrganizationMembershipToResponse(o struct {
	Role         Role
	Organization struct {
		ID   string
		Name string
		Slug string
	}
},
) string {
	return fmt.Sprintf(`{
		"role": "%s",
		"organization": {
			"id": "%s",
			"name": "%s",
			"slug": "%s"
		}
	}`, string(o.Role), o.Organization.ID, o.Organization.Name, o.Organization.Slug)
}

func DeleteSuccessQueryResult() (string, string) {
	resultTypename := "ToolDeleteSuccess"

	return fmt.Sprintf(`{
		"data": {
			"%s": {
				"__typename": "%s",
				"result": "%s"
			}
		}
	}`, "toolDelete", resultTypename, "Success"), resultTypename
}

func DeleteFailureQueryResult(failure string) (string, string) {
	resultTypename := "ToolDeleteFailure"
	return fmt.Sprintf(`{
			"data": {
				"%s": {
					"__typename": "%s",
					"result": "%s"
				}
			}
		}`, "toolDelete", resultTypename, failure), resultTypename
}

func quoteStringOrNull(s string) string {
	if s != "" {
		return fmt.Sprintf(`"%s"`, s)
	}

	return "null"
}
