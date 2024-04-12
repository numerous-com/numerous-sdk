package push

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type DeployEvent struct {
	Typename string `json:"__typename"`
	Result   string `json:"result"`
}

type subscription3 struct {
	DeployEvents DeployEvent `json:"deployEvents"`
}

type subscription2 struct {
	DeployEvents struct {
		Typename string             `graphql:"__typename"`
		Success  *BuildEventSuccess `graphql:"... on BuildEventSuccess"`
		Failure  *BuildEventFailure `graphql:"... on BuildEventFailure"`
		Info     *BuildEventInfo    `graphql:"... on BuildEventInfo"`
	} `graphql:"deployEvents(toolID: $toolID)"`
}

func getDeployEventLogs(toolID string) error {
	client := getClient()
	defer client.Close()

	err := deployEventSubscription(client, toolID, true)
	if err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return err
}

func deployEventSubscription(client *graphql.SubscriptionClient, toolID string, verbose bool) error {
	var sub subscription2
	var out io.Writer = os.Stdout

	variables := map[string]any{"toolID": graphql.ID(toolID)}
	_, err := client.Subscribe(&sub, variables, func(dataValue []byte, err error) error {
		if err != nil {
			return err
		}
		data := subscription3{}
		if err := json.Unmarshal([]byte(dataValue), &data); err != nil {
			return err
		}
		if data.DeployEvents.Typename == "BuildEventInfo" {
			for _, msg := range strings.Split(data.DeployEvents.Result, "\r\n") {
				var b BuildEventMessage

				if err := json.Unmarshal([]byte(msg), &b); err != nil {
					slog.Debug("error unmarshalling build event message", slog.Any("message", msg), slog.Any("error", err))
					continue
				}

				if b.Message != "" {
					printVerbose(out, b.Message, verbose)
				}
			}
		}

		if data.DeployEvents.Typename == "BuildEventFailure" {
			ProcessBuildEvent(data.DeployEvents.Result, out, true)
			return nil
		}

		if data.DeployEvents.Typename == "BuildEventSuccess" {
			ProcessBuildEvent(data.DeployEvents.Result, out, true)
			return nil
		}

		return nil
	})

	return err
}
