package status

import (
	"context"
	"fmt"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/timeseries"
)

type statusInput struct {
	appDir       string
	appSlug      string
	orgSlug      string
	metricsSince *time.Time
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

	workloads, err := apps.ListAppWorkloads(ctx, app.ListAppWorkloadsInput{AppID: readOutput.AppID, MetricsSince: input.metricsSince})
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

	fmt.Printf("  Status: %s\n", workload.Status)
	fmt.Printf("  Started at: %s (up for %s)\n", workload.StartedAt.Format(time.DateTime), humanizeDuration(time.Since(workload.StartedAt)))
	printCPUUsage(workload.CPUUsage)
	printMemoryUsage(workload.MemoryUsageMB)
	printLogs(workload.LogEntries)
}

func printLogs(entries []app.AppDeployLogEntry) {
	fmt.Println("  Logs (last 10 lines):")
	for _, entry := range entries {
		fmt.Println("    "+output.AnsiFaint+entry.Timestamp.Format(time.RFC3339)+output.AnsiReset, entry.Text)
	}
}

func printCPUUsage(cpuUsage app.AppWorkloadResourceUsage) {
	fmt.Printf("  CPU Usage (1024·vCPU): %s\n", formatUsage(cpuUsage))
	printPlot("    ", cpuUsage.Timeseries)
}

func printMemoryUsage(memoryUsageMB app.AppWorkloadResourceUsage) {
	fmt.Printf("  Memory Usage (MB): %s\n", formatUsage(memoryUsageMB))
	printPlot("    ", memoryUsageMB.Timeseries)
}

func printPlot(prefix string, t timeseries.Timeseries) {
	if len(t) == 0 {
		return
	}

	plotHeight := 10
	p := output.NewPlot(t)
	p.Display(prefix, plotHeight)
}

func formatUsage(usage app.AppWorkloadResourceUsage) string {
	if usage.Limit == nil {
		return fmt.Sprintf("%2.f", usage.Current)
	}

	return fmt.Sprintf("%2.f / %2.f", usage.Current, *usage.Limit)
}
