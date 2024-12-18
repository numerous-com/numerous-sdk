package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/timeseries"
)

const plotUpwardsDiagonalLine = `100.00┐                                       ██
      │                                    █████
      │                                █████████
      │                            █████████████
      │                       ██████████████████
      │                  ███████████████████████
      │             ████████████████████████████
      │        █████████████████████████████████
      │   ██████████████████████████████████████
      │█████████████████████████████████████████
  0.00┼────────────────────────────────────────┐
      2024-01-01T12:00:01Z  2024-01-01T12:00:05Z
`

const plotDownwardsDiagonalLine = `100.00┐███                                     
      │███████                                 
      │███████████                             
      │████████████████                        
      │████████████████████                    
      │████████████████████████                
      │█████████████████████████████           
      │█████████████████████████████████       
      │█████████████████████████████████████   
      │████████████████████████████████████████
  0.00┼───────────────────────────────────────┐
      2024-01-01T12:00:00Z 2024-01-02T14:00:00Z
`

const plotPointy = `30.00┐                  ███                   
     │                ███████                 
     │              ███████████               
     │            ███████████████             
     │          ████████████████████          
     │        ████████████████████████        
     │      ████████████████████████████      
     │    ████████████████████████████████    
     │  ████████████████████████████████████  
     │████████████████████████████████████████
 0.00┼───────────────────────────────────────┐
     2024-01-01T12:00:00Z 2024-01-01T14:00:00Z
`

const plotFlat = `31.00┐                                         
     │                                         
     │                                         
     │                                         
     │                                         
     │█████████████████████████████████████████
     │█████████████████████████████████████████
     │█████████████████████████████████████████
     │█████████████████████████████████████████
     │█████████████████████████████████████████
29.00┼────────────────────────────────────────┐
     2024-01-01T12:00:00Z  2024-01-01T14:00:00Z
`

const plotPartialXLabel = `31.00┐                                  
     │                                  
     │                                  
     │                                  
     │                                  
     │██████████████████████████████████
     │██████████████████████████████████
     │██████████████████████████████████
     │██████████████████████████████████
     │██████████████████████████████████
29.00┼─────────────────────────────────┐
     2024-01-01T12:0.. ..01-01T14:00:00Z
`

const plotNoXLabel = `31.00┐         
     │         
     │         
     │         
     │         
     │█████████
     │█████████
     │█████████
     │█████████
     │█████████
29.00┴─────────
`

func TestDisplay(t *testing.T) {
	type testCase struct {
		name       string
		expected   string
		timeseries timeseries.Timeseries
		termWidth  int
	}
	now := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)

	for _, tc := range []testCase{
		{
			name:     "upwards diagonal line",
			expected: plotUpwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 0.0, Timestamp: now.Add(time.Second)},
				{Value: 25.0, Timestamp: now.Add(2 * time.Second)},
				{Value: 50.0, Timestamp: now.Add(3 * time.Second)},
				{Value: 75.0, Timestamp: now.Add(4 * time.Second)},
				{Value: 100.0, Timestamp: now.Add(5 * time.Second)},
			},
			termWidth: 48,
		},
		{
			name:     "downwards diagonal line",
			expected: plotDownwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 100.0, Timestamp: now},
				{Value: 0.0, Timestamp: now.Add(time.Hour * 26)},
			},
			termWidth: 47,
		},
		{
			name:      "plots pointy",
			expected:  plotPointy,
			termWidth: 46,
			timeseries: timeseries.Timeseries{
				{Value: 0.0, Timestamp: now},
				{Value: 30.0, Timestamp: now.Add(time.Hour)},
				{Value: 0.0, Timestamp: now.Add(time.Hour).Add(time.Hour)},
			},
		},
		{
			name:      "flat timeseries is shown in the middle",
			expected:  plotFlat,
			termWidth: 47,
			timeseries: timeseries.Timeseries{
				{Value: 30.0, Timestamp: now},
				{Value: 30.0, Timestamp: now.Add(time.Hour)},
				{Value: 30.0, Timestamp: now.Add(time.Hour).Add(time.Hour)},
			},
		},
		{
			name:      "x-label is partially displayed if terminal is too narrow",
			expected:  plotPartialXLabel,
			termWidth: 40,
			timeseries: timeseries.Timeseries{
				{Value: 30.0, Timestamp: now},
				{Value: 30.0, Timestamp: now.Add(time.Hour)},
				{Value: 30.0, Timestamp: now.Add(time.Hour).Add(time.Hour)},
			},
		},
		{
			name:      "x-label is omitted if terminal is too narrow",
			expected:  plotNoXLabel,
			termWidth: 15,
			timeseries: timeseries.Timeseries{
				{Value: 30.0, Timestamp: now},
				{Value: 30.0, Timestamp: now.Add(time.Hour)},
				{Value: 30.0, Timestamp: now.Add(time.Hour).Add(time.Hour)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			stubTerm := &stubTerminal{width: tc.termWidth, buf: buf}

			p := NewPlotWithTerm(tc.timeseries, stubTerm)
			p.Display("", 10)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
