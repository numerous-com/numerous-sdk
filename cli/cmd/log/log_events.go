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

type LogsSubscription struct {
	Logs LogEntry `graphql:"logs(appId: $appId)"`
}

type SubscriptionClient interface {
	Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error)
	Run() error
	Close() error
}

func getLogs(appID string, timestamp bool) error {
	client := getClient()
	defer client.Close()
	err := logsSubscription(client, os.Stdout, appID, timestamp, true)
	if err != nil {
		return err
	}

	return err
}

func logsSubscription(client SubscriptionClient, out io.Writer, appID string, timestamps, verbose bool) error {
	var sub LogsSubscription
	variables := map[string]any{"appId": graphql.ID(appID)}

	_, err := client.Subscribe(&sub, variables, func(dataValue []byte, err error) error {
		// println("got logs event:", string(dataValue), ", ", err)
		if err != nil {
			return err
		}
		var data LogsSubscription

		if err := json.Unmarshal(dataValue, &data); err != nil {
			return err
		}
		processLogEntry(data, out, timestamps, verbose)

		return nil
	})
	if err != nil {
		return err
	}

	return client.Run()
}

func processLogEntry(entry LogsSubscription, out io.Writer, timestamps, verbose bool) {
	if entry.Logs.Message != "" {
		printVerbose(out, entry.Logs, timestamps, verbose)
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
		slog.Error("Error writing message", slog.String("error", err.Error()))
	}
}

func getClient() SubscriptionClient {
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
