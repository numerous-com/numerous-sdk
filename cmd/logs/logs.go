package logs

import (
	"context"
	"fmt"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
)

type appService interface {
	AppDeployLogs(appident.AppIdentifier) (chan app.AppDeployLogEntry, error)
}

type logsInput struct {
	appDir  string
	orgSlug string
	appSlug string
	printer func(app.AppDeployLogEntry)
}

func logs(ctx context.Context, apps appService, input logsInput) error {
	ai, err := appident.GetAppIdentifier(input.appDir, nil, input.orgSlug, input.appSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, input.appDir, ai)
		return err
	}

	ch, err := apps.AppDeployLogs(ai)
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return nil
			}
			input.printer(entry)
		case <-ctx.Done():
			return nil
		}
	}
}

func TimestampPrinter(entry app.AppDeployLogEntry) {
	ts := output.AnsiFaint + entry.Timestamp.Format(time.RFC3339) + output.AnsiReset
	fmt.Println(ts + " " + entry.Text)
}

func TextPrinter(entry app.AppDeployLogEntry) {
	fmt.Println(entry.Text)
}
