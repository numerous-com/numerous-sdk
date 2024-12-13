package timeseries

import (
	"math"
	"time"
)

type TimeseriesPoint struct {
	Value     float64
	Timestamp time.Time
}

type Timeseries []TimeseriesPoint

func (t Timeseries) MaxValue() float64 {
	if len(t) == 0 {
		return 0.0
	}

	maxVal := t[0].Value
	for _, p := range t {
		maxVal = math.Max(maxVal, p.Value)
	}

	return maxVal
}

func (t Timeseries) MinValue() float64 {
	if len(t) == 0 {
		return 0.0
	}

	minVal := t[0].Value
	for _, p := range t {
		minVal = math.Min(minVal, p.Value)
	}

	return minVal
}

func (t Timeseries) MaxTimestamp() time.Time {
	if len(t) == 0 {
		return time.Unix(0, 0)
	}

	maxTS := t[0].Timestamp
	for _, p := range t {
		if p.Timestamp.After(maxTS) {
			maxTS = p.Timestamp
		}
	}

	return maxTS
}

func (t Timeseries) MinTimestamp() time.Time {
	if len(t) == 0 {
		return time.Unix(0, 0)
	}

	minTS := t[0].Timestamp
	for _, p := range t {
		if p.Timestamp.Before(minTS) {
			minTS = p.Timestamp
		}
	}

	return minTS
}
