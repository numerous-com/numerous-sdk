package app

import (
	"context"
	"errors"
	"time"
)

type ListApp struct {
	Name        string
	Slug        string
	Description string
	Status      string
	CreatedBy   string
	CreatedAt   time.Time
}

type QueryApp struct {
	DisplayName string
	Slug        string
	Description string
	CreatedBy   struct {
		FullName string
	}
	CreatedAt         time.Time
	DefaultDeployment *struct {
		Current *struct {
			Status string
		}
	}
}

func (qa QueryApp) ToListApp() ListApp {
	status := "NOT DEPLOYED"
	if qa.DefaultDeployment != nil && qa.DefaultDeployment.Current != nil {
		status = qa.DefaultDeployment.Current.Status
	}

	return ListApp{
		Name:        qa.DisplayName,
		Slug:        qa.Slug,
		Description: qa.Description,
		Status:      status,
		CreatedBy:   qa.CreatedBy.FullName,
		CreatedAt:   qa.CreatedAt,
	}
}

type QueryOrganizationApps struct {
	OrganizationApps struct {
		Typename     string `graphql:"__typename"`
		Organization struct {
			Apps []QueryApp
		} `graphql:"... on Organization"`
	} `graphql:"organization(organizationSlug: $organizationSlug)"`
}

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUnexpectedType       = errors.New("unexpected type")
)

func (s *Service) List(ctx context.Context, organizationSlug string) ([]ListApp, error) {
	var q QueryOrganizationApps
	if err := s.client.Query(ctx, &q, map[string]interface{}{"organizationSlug": organizationSlug}); err != nil {
		return nil, convertErrors(err)
	}

	if q.OrganizationApps.Typename == "OrganizationNotFound" {
		return nil, ErrOrganizationNotFound
	} else if q.OrganizationApps.Typename != "Organization" {
		return nil, ErrUnexpectedType
	}

	apps := []ListApp{}
	for _, qa := range q.OrganizationApps.Organization.Apps {
		apps = append(apps, qa.ToListApp())
	}

	return apps, nil
}