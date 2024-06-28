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

func Logs(ctx context.Context, apps AppService, appDir, slug, appSlug string, printer func(app.AppDeployLogEntry)) error {
	ai, err := appident.GetAppIdentifier(appDir, slug, appSlug)
	if err != nil {
		return err
	}

	ch, err := apps.AppDeployLogs(ai)
	if err != nil {
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
