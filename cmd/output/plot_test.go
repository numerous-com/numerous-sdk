package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/timeseries"
)

const plotUpwardsDiagonalLine = `         #
        # 
       #  
      #   
     #    
    #     
   #      
  #       
 #        
#         
`

const plotDownwardsDiagonalLine = `#         
 #        
  #       
   #      
    #     
     #    
      #   
       #  
        # 
         #
`

const plotPointy = `    #     
    ##    
   # #    
   #  #   
  #   #   
  #    #  
 #     #  
 #      # 
#       # 
#        #
`

func TestDisplay(t *testing.T) {
	type testCase struct {
		name       string
		expected   string
		timeseries timeseries.Timeseries
	}
	now := time.Date(2024, time.January, 1, 1, 1, 0, 0, time.UTC)

	for _, tc := range []testCase{
		{
			name:     "upwards diagonal line from multiple points",
			expected: plotUpwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 0.0, Timestamp: now.Add(time.Second)},
				{Value: 10.0, Timestamp: now.Add(2 * time.Second)},
				{Value: 20.0, Timestamp: now.Add(3 * time.Second)},
				{Value: 30.0, Timestamp: now.Add(4 * time.Second)},
			},
		},
		{
			name:     "upwards diagonal line between two points",
			expected: plotUpwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 1000.0, Timestamp: now.Add(time.Second)},
				{Value: 100000.0, Timestamp: now.Add(time.Hour * 123)},
			},
		},
		{
			name:     "downwards diagonal line from multiple points",
			expected: plotDownwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 30.0, Timestamp: now.Add(time.Second)},
				{Value: 20.0, Timestamp: now.Add(2 * time.Second)},
				{Value: 10.0, Timestamp: now.Add(3 * time.Second)},
				{Value: 0.0, Timestamp: now.Add(4 * time.Second)},
			},
		},
		{
			name:     "downwards diagonal line between two points",
			expected: plotDownwardsDiagonalLine,
			timeseries: timeseries.Timeseries{
				{Value: 100000.0, Timestamp: now.Add(time.Second)},
				{Value: 1000.0, Timestamp: now.Add(time.Hour * 123)},
			},
		},
		{
			name:     "plots pointy",
			expected: plotPointy,
			timeseries: timeseries.Timeseries{
				{Value: 0.0, Timestamp: now},
				{Value: 10.0, Timestamp: now.Add(time.Hour)},
				{Value: 0.0, Timestamp: now.Add(time.Hour).Add(time.Hour)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			stubTerm := &stubTerminal{width: 10, buf: buf}

			p := NewPlotWithTerm(tc.timeseries, stubTerm)
			p.Display("", 10)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
