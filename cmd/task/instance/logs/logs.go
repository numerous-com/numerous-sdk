package logs

import (
	"context"
	"fmt"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/output"
)

type taskLogsService interface {
	TaskInstanceLogs(input app.TaskInstanceLogsInput) (chan app.WorkloadLogEntry, error)
}

type taskLogsInput struct {
	instanceID string
	tail       int
	follow     bool
	printer    func(app.WorkloadLogEntry)
}

func taskLogs(ctx context.Context, service taskLogsService, input taskLogsInput) error {
	var tail *int
	if input.tail > 0 {
		tail = &input.tail
	}

	serviceInput := app.TaskInstanceLogsInput{
		InstanceID: input.instanceID,
		Tail:       tail,
		Follow:     input.follow,
	}

	ch, err := service.TaskInstanceLogs(serviceInput)
	if err != nil {
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

func TimestampPrinter(entry app.WorkloadLogEntry) {
	ts := output.AnsiFaint + entry.Timestamp.Format(time.RFC3339) + output.AnsiReset
	fmt.Println(ts + " " + entry.Text)
}

func TextPrinter(entry app.WorkloadLogEntry) {
	fmt.Println(entry.Text)
}
