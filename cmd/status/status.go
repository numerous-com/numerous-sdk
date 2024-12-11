package status

import (
	"context"
	"fmt"
	"math"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
)

type statusInput struct {
	appDir  string
	appSlug string
	orgSlug string
}

type appReaderWorkloadLister interface {
	ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error)
	ListAppWorkloads(ctx context.Context, input app.ListAppWorkloadsInput) ([]app.AppWorkload, error)
}

func status(ctx context.Context, apps appReaderWorkloadLister, input statusInput) error {
	ai, err := appident.GetAppIdentifier(input.appDir, nil, input.orgSlug, input.appSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, input.appDir, ai)
		return err
	}

	readOutput, err := apps.ReadApp(ctx, app.ReadAppInput{OrganizationSlug: ai.OrganizationSlug, AppSlug: ai.AppSlug})
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	workloads, err := apps.ListAppWorkloads(ctx, app.ListAppWorkloadsInput(readOutput))
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	printWorkloads(workloads)

	return nil
}

func printWorkloads(workloads []app.AppWorkload) {
	first := true
	for _, w := range workloads {
		if !first {
			println()
		}
		printWorkload(w)
		first = false
	}
}

func printWorkload(workload app.AppWorkload) {
	if workload.OrganizationSlug != "" {
		fmt.Printf("Workload in %q:\n", workload.OrganizationSlug)
	} else if sub := workload.Subscription; sub != nil {
		fmt.Printf("Workload for subscription %q in %q:\n", sub.SubscriptionUUID, sub.OrganizationSlug)
	} else {
		fmt.Println("Workload of unknown origin:")
	}

	fmt.Printf("    Status: %s\n", workload.Status)
	fmt.Printf("    Started at: %s (up for %s)\n", workload.StartedAt.Format(time.DateTime), humanDuration(time.Since(workload.StartedAt)))
}

const (
	hoursPerDay      float64 = 24.0
	minutesPerHour   float64 = 60.0
	secondsPerMinute float64 = 60.0
)

func humanDuration(since time.Duration) string {
	hours := since.Hours()
	if hours > hoursPerDay {
		fullDays := math.Floor(hours / hoursPerDay)
		dayHours := math.Floor(hours - fullDays*hoursPerDay)
		if dayHours > 0.0 {
			return fmt.Sprintf("%d days and %d hours", int(fullDays), int(dayHours))
		} else {
			return fmt.Sprintf("%d days", int(fullDays))
		}
	}

	minutes := since.Minutes()
	if hours > 1.0 {
		fullHours := math.Floor(hours)
		hourMinutes := minutes - fullHours*minutesPerHour
		if fullHours > 0.0 {
			return fmt.Sprintf("%d hours and %d minutes", int(fullHours), int(hourMinutes))
		}
	}

	seconds := since.Seconds()
	if minutes > 1.0 {
		fullMinutes := math.Floor(minutes)
		minuteSeconds := seconds - fullMinutes*secondsPerMinute
		if fullMinutes > 0.0 {
			return fmt.Sprintf("%d minutes and %d seconds", int(fullMinutes), int(minuteSeconds))
		}
	}

	return fmt.Sprintf("%d seconds", int(math.Round(seconds)))
}
