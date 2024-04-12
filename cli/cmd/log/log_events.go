package log

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

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
			fmt.Printf("Error occurred listening for deploy logs. This does not mean that you app will be unavailable.\nFirst error: %s\nSecond error: %s\n", previousError, err)
			return err
		}
		fmt.Printf("Error occurred listening for deploy logs.\nError: %s\nRetrying...\n", err)
		previousError = err

		return nil
	})

	return client
}

func getLogs(appID string) error {
	client := getClient()
	defer client.Close()
	err := logsSubscription(client, appID, true)
	if err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return err
}

func logsSubscription(client *graphql.SubscriptionClient, appID string, verbose bool) error {
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
		ProcessLogEntry(data, out, verbose)

		return nil
	})

	return err
}

func ProcessLogEntry(entry LogsContainer, out io.Writer, verbose bool) {
	if entry.Logs.Message != "" {
		printVerbose(out, entry.Logs.Message, verbose)
	}
}

func printVerbose(out io.Writer, message string, verbose bool) {
	if verbose {
		_, err := out.Write([]byte(message + "\n"))
		if err != nil {
			slog.Error("Error writing message", err)
		}
	}
}
