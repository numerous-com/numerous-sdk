package status

import (
	"fmt"
	"math"
	"time"
)

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
