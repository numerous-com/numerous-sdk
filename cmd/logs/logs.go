package logs

import (
	"context"
	"fmt"
	"time"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type AppService interface {
	AppDeployLogs(appident.AppIdentifier) (chan app.AppDeployLogEntry, error)
}

func Logs(ctx context.Context, apps AppService, appDir, orgSlug, appSlug string, printer func(app.AppDeployLogEntry)) error {
	ai, err := appident.GetAppIdentifier(appDir, nil, orgSlug, appSlug)
	if err != nil {
		appident.PrintGetAppIdentiferError(err, appDir, ai)
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
			printer(entry)
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
