package app

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	"numerous.com/cli/internal/timeseries"
)

type AppWorkloadSubscription struct {
	OrganizationSlug string
	SubscriptionUUID string
}

type AppWorkloadResourceUsage struct {
	Current    float64
	Limit      *float64
	Timeseries timeseries.Timeseries
}

type AppWorkload struct {
	OrganizationSlug string
	Subscription     *AppWorkloadSubscription
	StartedAt        time.Time
	Status           string
	LogEntries       []AppDeployLogEntry
	CPUUsage         AppWorkloadResourceUsage
	MemoryUsageMB    AppWorkloadResourceUsage
}

type ListAppWorkloadsInput struct {
	AppID        string
	MetricsSince *time.Time
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
	Logs      struct {
		Edges []AppDeployLogEntry
	}
	CPUUsage      AppWorkloadResourceUsage
	MemoryUsageMB AppWorkloadResourceUsage
}

type appWorkloadsResponse struct {
	AppWorkloads []appWorkloadsResponseWorkload `graphql:"appWorkloads(appID: $appID)"`
}

const queryAppWorkloadsText = `
query CLIListAppWorkloads($appID: ID!, $metricsSince: Time!) {
	appWorkloads(appID: $appID, input: {metricsSince: $metricsSince}) {
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
		logs(last: 10) {
			edges {
				timestamp
				text
			}
		}
		cpuUsage {
			current
			limit
			timeseries {
				timestamp
				value
			}
		}
		memoryUsageMB {
			current
			limit
			timeseries {
				timestamp
				value
			}
		}
	}
}
`

func (s *Service) ListAppWorkloads(ctx context.Context, input ListAppWorkloadsInput) ([]AppWorkload, error) {
	var metricsSince time.Time
	if input.MetricsSince != nil {
		metricsSince = *input.MetricsSince
	} else {
		// If this default is changed, remember to update the help text default text for the `--metrics-since` flag.
		metricsSince = s.clock.Now().Add(-time.Hour)
	}

	var resp appWorkloadsResponse
	variables := map[string]any{"appID": input.AppID, "metricsSince": metricsSince.Format(time.RFC3339)}

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
		StartedAt:     responseWorkload.StartedAt,
		Status:        responseWorkload.Status,
		LogEntries:    responseWorkload.Logs.Edges,
		CPUUsage:      responseWorkload.CPUUsage,
		MemoryUsageMB: responseWorkload.MemoryUsageMB,
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
