package push

import (
	"encoding/json"
	"io"
	"log/slog"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type DeployEvent struct {
	Typename string             `graphql:"__typename"`
	Success  *BuildEventSuccess `graphql:"... on BuildEventSuccess"`
	Failure  *BuildEventFailure `graphql:"... on BuildEventFailure"`
	Info     *BuildEventInfo    `graphql:"... on BuildEventInfo"`
}

type DeployEventsSubscription struct {
	DeployEvents DeployEvent `graphql:"deployEvents(toolID: $toolID)"`
}

func getDeployEventLogs(w io.Writer, toolID string) error {
	client := getClient()
	defer client.Close()

	err := deployEventSubscription(client, w, toolID, true)
	if err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return err
}

func deployEventSubscription(client *graphql.SubscriptionClient, w io.Writer, toolID string, verbose bool) error {
	variables := map[string]any{"toolID": graphql.ID(toolID)}
	_, err := client.Subscribe(&DeployEventsSubscription{}, variables, func(dataValue []byte, err error) error {
		if err != nil {
			return err
		}

		deployEvent := DeployEventsSubscription{}
		if err := json.Unmarshal([]byte(dataValue), &deployEvent); err != nil {
			return err
		}

		if deployEvent.DeployEvents.Typename == "BuildEventInfo" {
			for _, msg := range strings.Split(deployEvent.DeployEvents.Info.Result, "\r\n") {
				var b BuildEventMessage

				if err := json.Unmarshal([]byte(msg), &b); err != nil {
					slog.Debug("error unmarshalling build event message", slog.Any("message", msg), slog.Any("error", err))
					continue
				}

				if b.Message != "" {
					printVerbose(w, b.Message, verbose)
				}
			}
		}

		if deployEvent.DeployEvents.Typename == "BuildEventFailure" {
			ProcessBuildEvent(w, deployEvent.DeployEvents.Failure.Result, true)
			return nil
		}

		if deployEvent.DeployEvents.Typename == "BuildEventSuccess" {
			ProcessBuildEvent(w, deployEvent.DeployEvents.Success.Result, true)
			return nil
		}

		return nil
	})

	return err
}
