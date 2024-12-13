package status

import (
	"context"
	"fmt"
	"math"
	"time"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/timeseries"
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

	println("Name: " + readOutput.AppDisplayName)
	if readOutput.AppDescription != "" {
		println("Description: " + readOutput.AppDescription)
	}

	workloads, err := apps.ListAppWorkloads(ctx, app.ListAppWorkloadsInput{AppID: readOutput.AppID})
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	println()
	if len(workloads) == 0 {
		println("No workloads found")
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
	fmt.Printf("    Started at: %s (up for %s)\n", workload.StartedAt.Format(time.DateTime), humanizeDuration(time.Since(workload.StartedAt)))
	printCPUUsage(workload.CPUUsage)
	printMemoryUsage(workload.MemoryUsageMB)
	printLogs(workload.LogEntries)
}

func printLogs(entries []app.AppDeployLogEntry) {
	fmt.Println("    Logs (last 10 lines):")
	for _, entry := range entries {
		fmt.Println("        ", output.AnsiFaint, entry.Timestamp.Format(time.RFC3339), output.AnsiReset, entry.Text)
	}
}

func printCPUUsage(cpuUsage app.AppWorkloadResourceUsage) {
	fmt.Printf("    CPU Usage: %s\n", formatUsage(cpuUsage))
	printPlot("        |", cpuUsage.Timeseries)
}

func printMemoryUsage(memoryUsageMB app.AppWorkloadResourceUsage) {
	fmt.Printf("    Memory Usage (MB): %s\n", formatUsage(memoryUsageMB))
	printPlot("        |", memoryUsageMB.Timeseries)
}

func printPlot(prefix string, t timeseries.Timeseries) {
	if len(t) == 0 {
		return
	}

	plotHeight := 10
	p := output.NewPlot(t)
	p.Display(prefix, plotHeight)
}

const (
	hoursPerDay      int = 24
	minutesPerHour   int = 60
	secondsPerMinute int = 60
)

func humanizeDuration(since time.Duration) string {
	hours := int(math.Floor(since.Hours()))
	if hours > hoursPerDay {
		fullDays := hours / hoursPerDay
		remainingHours := hours % hoursPerDay
		if remainingHours > 0 {
			return fmt.Sprintf("%d days and %d hours", fullDays, remainingHours)
		} else {
			return fmt.Sprintf("%d days", fullDays)
		}
	}

	minutes := int(math.Floor(since.Minutes()))
	if hours > 1 {
		fullHours := hours
		remainingMinutes := minutes % minutesPerHour
		if fullHours > 0 {
			return fmt.Sprintf("%d hours and %d minutes", fullHours, remainingMinutes)
		}
	}

	seconds := int(math.Round(since.Seconds()))
	if minutes > 1 {
		fullMinutes := minutes
		remainingSeconds := seconds % secondsPerMinute
		if fullMinutes > 0.0 {
			return fmt.Sprintf("%d minutes and %d seconds", fullMinutes, remainingSeconds)
		}
	}

	return fmt.Sprintf("%d seconds", seconds)
}

func formatUsage(usage app.AppWorkloadResourceUsage) string {
	if usage.Limit == nil {
		return fmt.Sprintf("%2.f", usage.Current)
	}

	return fmt.Sprintf("%2.f / %2.f", usage.Current, *usage.Limit)
}
