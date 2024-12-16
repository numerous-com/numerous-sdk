package timeseries

import "time"

type NormalizedPoint struct {
	X float64
	Y float64
}

type NormalizedTimeseries []NormalizedPoint

func (t Timeseries) Normalize(xScale float64, yScale float64) NormalizedTimeseries {
	minVal := t.MinValue()
	maxVal := t.MaxValue()
	valDiff := maxVal - minVal
	minTS := t.MinTimestamp()
	maxTS := t.MaxTimestamp()

	normalizeTS := func(ts time.Time) float64 {
		ratio := float64(ts.Sub(minTS).Microseconds()) / float64(maxTS.Sub(minTS).Microseconds())
		return xScale * ratio
	}

	var normalizeValue func(v float64) float64
	if valDiff != 0.0 {
		normalizeValue = func(v float64) float64 {
			ratio := (v - minVal) / valDiff
			return yScale * ratio
		}
	} else {
		// "vertically" center data with no "vertical" differences
		normalizeValue = func(float64) float64 {
			return yScale * 0.5 // nolint:mnd
		}
	}

	var normalized NormalizedTimeseries
	for _, p := range t {
		normalized = append(normalized, NormalizedPoint{
			X: normalizeTS(p.Timestamp),
			Y: normalizeValue(p.Value),
		})
	}

	return normalized
}
