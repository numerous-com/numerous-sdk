package log

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"

	"numerous/cli/cmd/output"
	"numerous/cli/internal/gql"

	"github.com/hasura/go-graphql-client"
)

type LogEntry struct {
	Time    time.Time `json:"time,omitempty"`
	Message string    `json:"message,omitempty"`
}

type LogsContainer struct {
	Logs LogEntry `json:"logs"`
}

type subscription struct {
	LogMessage LogEntry `graphql:"logs(appId: $appId)"`
}

func getClient() *graphql.SubscriptionClient {
	var previousError error
	client := gql.GetSubscriptionClient()
	client = client.OnError(func(sc *graphql.SubscriptionClient, err error) error {
		if previousError != nil {
			output.PrintError(
				"Error occurred subscribing to app logs",
				"This does not mean that you app will be unavailable.\n"+
					"First error: %s\n"+
					"Second error: %s\n",
				previousError, err,
			)

			return err
		}

		output.PrintErrorDetails("Error occurred subscribing to app logs. Retrying...", err)
		previousError = err

		return nil
	})

	return client
}

func getLogs(appID string, timestamp bool) error {
	client := getClient()
	defer client.Close()
	err := logsSubscription(client, appID, timestamp, true)
	if err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return err
}

func logsSubscription(client *graphql.SubscriptionClient, appID string, timestamps, verbose bool) error {
	var sub subscription
	out := os.Stdout
	variables := map[string]any{"appId": graphql.ID(appID)}

	_, err := client.Subscribe(&sub, variables, func(dataValue []byte, err error) error {
		if err != nil {
			return err
		}
		var data LogsContainer

		if err := json.Unmarshal(dataValue, &data); err != nil {
			return err
		}
		ProcessLogEntry(data, out, timestamps, verbose)

		return nil
	})

	return err
}

func ProcessLogEntry(entry LogsContainer, out io.Writer, timestamps, verbose bool) {
	if entry.Logs.Message != "" {
		printVerbose(out, entry.Logs, verbose, timestamps)
	}
}

func printVerbose(out io.Writer, entry LogEntry, timestamps, verbose bool) {
	if !verbose {
		return
	}

	logMsg := entry.Message + "\n"
	if timestamps {
		logMsg = entry.Time.Format(time.RFC3339) + " " + logMsg
	}

	if _, err := out.Write([]byte(logMsg)); err != nil {
		slog.Error("Error writing message", err)
	}
}
