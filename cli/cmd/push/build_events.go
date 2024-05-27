package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"numerous/cli/internal/gql"

	"github.com/hasura/go-graphql-client"
)

type BuildEvent struct {
	Typename string `json:"__typename"`
	Result   string `json:"result"`
}

type subscription struct {
	BuildEvents BuildEvent `json:"buildEvents"`
}

type BuildEventSuccess struct {
	Result string `json:"result"`
}

type BuildEventFailure struct {
	Result string `json:"result"`
}

type BuildEventInfo struct {
	Result string `json:"result"`
}

type subscription1 struct {
	BuildEvents struct {
		Typename string             `graphql:"__typename"`
		Success  *BuildEventSuccess `graphql:"... on BuildEventSuccess"`
		Failure  *BuildEventFailure `graphql:"... on BuildEventFailure"`
		Info     *BuildEventInfo    `graphql:"... on BuildEventInfo"`
	} `graphql:"buildEvents(buildId: $buildId, appPath: $appPath)"`
}

func getBuildEventLogs(buildID string, appPath string, verbose bool) error {
	client := getClient()
	defer client.Close()

	err := buildEventSubscription(client, buildID, appPath, verbose)
	if err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return err
}

type BuildEventErrorDetail struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BuildEventMessage struct {
	Status  string                `json:"status,omitempty"`
	Message string                `json:"stream,omitempty"`
	Error   BuildEventErrorDetail `json:"errorDetail,omitempty"`
}

type subscriptionClient interface {
	Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error)
}

func buildEventSubscription(client subscriptionClient, buildID string, appPath string, verbose bool) error {
	var sub subscription1
	var out io.Writer = os.Stdout

	variables := map[string]any{"buildId": graphql.ID(buildID), "appPath": graphql.String(appPath)}
	_, err := client.Subscribe(&sub, variables, func(dataValue []byte, err error) error {
		if err != nil {
			return err
		}
		data := subscription{}
		if err := json.Unmarshal([]byte(dataValue), &data); err != nil {
			return err
		}
		if data.BuildEvents.Typename == "BuildEventInfo" {
			for _, msg := range strings.Split(data.BuildEvents.Result, "\r\n") {
				ProcessBuildEvent(msg, out, verbose)
			}
		}

		if data.BuildEvents.Typename == "BuildEventFailure" {
			fmt.Fprintf(out, data.BuildEvents.Result)
			return errors.New(data.BuildEvents.Result)
		}

		return nil
	})

	return err
}

func getClient() *graphql.SubscriptionClient {
	var previousError error

	client := gql.NewSubscriptionClient()
	client = client.OnError(func(sc *graphql.SubscriptionClient, err error) error {
		if previousError != nil {
			fmt.Printf("Error occurred listening for deploy logs. This does not mean that you app will be unavailable.\nFirst error: %s\nSecond error: %s\n", previousError, err)
			return err
		}
		fmt.Printf("Error occurred listening for deploy logs.\nError: %s\nRetrying...\n", err)
		previousError = err

		return nil
	})

	return client
}

func ProcessBuildEvent(msg string, out io.Writer, verbose bool) {
	var b BuildEventMessage
	if err := json.Unmarshal([]byte(msg), &b); err != nil {
		slog.Debug("error unmarshalling build event message", slog.Any("message", msg), slog.Any("error", err))
		return
	}

	if b.Message != "" {
		printVerbose(out, b.Message, verbose)
	}
	if b.Status != "" {
		printVerbose(out, b.Status, verbose)
	}
}

// printVerbose filters away buildEventMessages that do not interest the average user
func printVerbose(out io.Writer, message string, verbose bool) {
	if verbose {
		fmt.Fprintf(out, "     Build: %s\n", strings.TrimSpace(message))
	}
}
