package status

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

type Since time.Time

func (s *Since) String() string {
	t := time.Time(*s)
	return t.Format(time.RFC3339)
}

func (s *Since) Set(v string) error {
	since, err := parseSince(v, time.Now())
	if err != nil {
		return err
	}

	*s = *since

	return nil
}

func (*Since) Type() string {
	return "time"
}

func (s *Since) Time() *time.Time {
	if s == nil {
		return nil
	}

	if time.Time(*s).IsZero() {
		return nil
	}

	t := time.Time(*s)

	return &t
}

var errParseSince = errors.New("could not parse since value")

func parseSince(value string, now time.Time) (*Since, error) {
	// Try supported timestamp formats
	for _, format := range []string{time.RFC3339, time.DateOnly} {
		if parsed, err := time.Parse(format, value); err == nil {
			since := Since(parsed)
			return &since, nil
		}
	}

	// Try relative time specifications
	relTimePat := regexp.MustCompile(`^([1-9][0-9]*)(d|h|m|s)$`)
	matches := relTimePat.FindStringSubmatch(value)
	if len(matches) != 3 { // nolint:mnd
		return nil, errParseSince
	}

	matchedValue, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, errParseSince
	}

	unitDuration, err := sinceUnitDuration(matches[2])
	if err != nil {
		return nil, errParseSince
	}

	result := Since(now.Add(-time.Duration(matchedValue) * unitDuration))

	return &result, nil
}

func sinceUnitDuration(unit string) (time.Duration, error) {
	switch unit {
	case "d":
		return time.Hour * 24, nil // nolint:mnd
	case "h":
		return time.Hour, nil
	case "m":
		return time.Minute, nil
	case "s":
		return time.Second, nil
	default:
		return 0, errParseSince
	}
}
