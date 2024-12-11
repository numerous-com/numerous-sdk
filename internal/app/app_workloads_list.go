package app

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
)

type AppWorkloadSubscription struct {
	OrganizationSlug string
	SubscriptionUUID string
}

type AppWorkload struct {
	OrganizationSlug string
	Subscription     *AppWorkloadSubscription
	StartedAt        time.Time
	Status           string
}

type ListAppWorkloadsInput struct {
	AppID string
}

type appWorkloadsResponseWorkload struct {
	Status       string
	Organization *struct {
		Slug string
	}
	Subscription *struct {
		ID                  string
		InboundOrganization struct {
			Slug string
		}
	}
	StartedAt time.Time
}

type appWorkloadsResponse struct {
	AppWorkloads []appWorkloadsResponseWorkload `graphql:"appWorkloads(appID: $appID)"`
}

const queryAppWorkloadsText = `
query CLIListAppWorkloads($appID: ID!) {
	appWorkloads(appID: $appID) {
		status
		startedAt
		organization {
			slug
		}
		subscription {
			id
			inboundOrganization {
				slug
			}
		}
	}
}
`

func (s *Service) ListAppWorkloads(ctx context.Context, input ListAppWorkloadsInput) ([]AppWorkload, error) {
	var resp appWorkloadsResponse
	variables := map[string]any{"appID": input.AppID}

	err := s.client.Exec(ctx, queryAppWorkloadsText, &resp, variables, graphql.OperationName("CLIListAppWorkloads"))
	if err != nil {
		return nil, convertErrors(err)
	}

	workloads := []AppWorkload{}
	for _, wl := range resp.AppWorkloads {
		workloads = append(workloads, appWorkloadFromResponse(wl))
	}

	return workloads, nil
}

func appWorkloadFromResponse(responseWorkload appWorkloadsResponseWorkload) AppWorkload {
	wl := AppWorkload{
		StartedAt: responseWorkload.StartedAt,
		Status:    responseWorkload.Status,
	}

	if responseWorkload.Organization != nil {
		wl.OrganizationSlug = responseWorkload.Organization.Slug
	} else if responseWorkload.Subscription != nil {
		wl.Subscription = &AppWorkloadSubscription{
			OrganizationSlug: responseWorkload.Subscription.InboundOrganization.Slug,
			SubscriptionUUID: responseWorkload.Subscription.ID,
		}
	}

	return wl
}
