package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"numerous.com/cli/internal/gql"

	"github.com/hasura/go-graphql-client"
)

type BuildEventSuccess struct {
	Result string `json:"result"`
}

type BuildEventFailure struct {
	Result string `json:"result"`
}

type BuildEventInfo struct {
	Result string `json:"result"`
}

type BuildEvent struct {
	Typename string            `graphql:"__typename"`
	Success  BuildEventSuccess `graphql:"... on BuildEventSuccess"`
	Failure  BuildEventFailure `graphql:"... on BuildEventFailure"`
	Info     BuildEventInfo    `graphql:"... on BuildEventInfo"`
}

type BuildEventSubscription struct {
	BuildEvents BuildEvent `graphql:"buildEvents(buildId: $buildId, appPath: $appPath)"`
}

func getBuildEventLogs(w io.Writer, buildID string, appPath string, verbose bool) error {
	client := getClient()
	defer client.Close()

	err := buildEventSubscription(client, w, buildID, appPath, verbose)
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
}

type BuildEventMessage struct {
	Status  string                `json:"status,omitempty"`
	Message string                `json:"stream,omitempty"`
	Error   BuildEventErrorDetail `json:"errorDetail,omitempty"`
}

type subscriptionClient interface {
	Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error)
}

func buildEventSubscription(client subscriptionClient, w io.Writer, buildID string, appPath string, verbose bool) error {
	variables := map[string]any{"buildId": graphql.ID(buildID), "appPath": graphql.String(appPath)}
	_, err := client.Subscribe(&BuildEventSubscription{}, variables, func(dataValue []byte, err error) error {
		if err != nil {
			return err
		}

		sub := BuildEventSubscription{}
		if err := graphql.UnmarshalGraphQL(dataValue, &sub); err != nil {
			return err
		}

		if sub.BuildEvents.Typename == "BuildEventInfo" {
			for _, msg := range strings.Split(sub.BuildEvents.Info.Result, "\r\n") {
				ProcessBuildEvent(w, msg, verbose)
			}
		}

		if sub.BuildEvents.Typename == "BuildEventFailure" {
			fmt.Fprintf(w, sub.BuildEvents.Failure.Result)
			return errors.New(sub.BuildEvents.Failure.Result)
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

func ProcessBuildEvent(w io.Writer, msg string, verbose bool) {
	var b BuildEventMessage
	if err := json.Unmarshal([]byte(msg), &b); err != nil {
		slog.Debug("error unmarshalling build event message", slog.Any("message", msg), slog.Any("error", err))
		return
	}

	if b.Message != "" {
		printVerbose(w, b.Message, verbose)
	}

	if b.Status != "" {
		printVerbose(w, b.Status, verbose)
	}
}

// printVerbose filters away buildEventMessages that do not interest the average user
func printVerbose(w io.Writer, message string, verbose bool) {
	if verbose {
		w.Write([]byte(strings.TrimSpace(message))) // nolint:errcheck
	}
}
