package output

import (
	"fmt"
	"math"
	"strings"

	"numerous.com/cli/internal/timeseries"
)

const (
	plotRune             rune = '#'
	defaultPlotLineWidth int  = 80
	maxPlotLineWidth     int  = 120
)

type Plot struct {
	term terminal
	data timeseries.Timeseries
}

func NewPlot(data timeseries.Timeseries) *Plot {
	return NewPlotWithTerm(data, newTermTerminal())
}

func NewPlotWithTerm(data timeseries.Timeseries, term terminal) *Plot {
	return &Plot{
		term: term,
		data: data,
	}
}

func (p *Plot) Display(prefix string, height int) {
	cols, _, err := p.term.GetSize()
	if err != nil {
		cols = defaultPlotLineWidth
	}

	if cols > maxPlotLineWidth {
		cols = maxPlotLineWidth
	}

	cols -= len(prefix)
	normalized := p.data.Normalize(float64(cols-1), float64(height-1))

	r := newRaster(prefix, cols, height)

	var prev *timeseries.NormalizedPoint
	for _, cur := range normalized {
		if prev == nil {
			prev = ref(cur)
			continue
		}

		r.bresenhamLine(int(prev.X), int(prev.Y), int(cur.X), int(cur.Y))

		prev = ref(cur)
	}

	w := p.term.Writer()
	fmt.Fprint(w, r.String())
}

func newRaster(prefix string, width, height int) *raster {
	return &raster{
		data:   make([]rune, width*height),
		width:  width,
		height: height,
		prefix: prefix,
	}
}

type raster struct {
	data   []rune
	width  int
	height int
	prefix string
}

func (r *raster) String() string {
	lines := []string{}
	for y := r.height - 1; y >= 0; y-- {
		ln := ""
		for x := 0; x < r.width; x++ {
			pt := r.data[r.width*y+x]
			if pt == 0 {
				ln += " "
			} else {
				ln += string(pt)
			}
		}
		lines = append(lines, ln)
	}

	return r.prefix + strings.Join(lines, "\n"+r.prefix) + "\n"
}

func (r *raster) plot(x, y int, val rune) {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return
	}

	i := y*r.width + x
	r.data[i] = val
}

func (r *raster) bresenhamLow(x0, y0, x1, y1 int) {
	dx := x1 - x0
	dy := y1 - y0
	yi := 1
	if dy < 0 {
		yi = -1
		dy = -dy
	}

	D := (2 * dy) - dx // nolint:mnd
	y := y0
	for x := x0; x <= x1; x++ {
		r.plot(x, y, plotRune)
		if D > 0 {
			y += yi
			D += (2 * (dy - dx)) // nolint:mnd
		} else {
			D += 2 * dy // nolint:mnd
		}
	}
}

func (r *raster) bresenhamHigh(x0, y0, x1, y1 int) {
	dx := x1 - x0
	dy := y1 - y0
	xi := 1
	if dx < 0 {
		xi = -1
		dx = -dx
	}
	D := (2 * dx) - dy // nolint:mnd
	x := x0

	for y := y0; y <= y1; y++ {
		r.plot(x, y, plotRune)
		if D > 0 {
			x += xi
			D += (2 * (dx - dy)) // nolint:mnd
		} else {
			D += 2 * dx // nolint:mnd
		}
	}
}

func (r *raster) bresenhamLine(x0, y0, x1, y1 int) {
	if math.Abs(float64(y1-y0)) < math.Abs(float64(x1-x0)) {
		if x0 > x1 {
			r.bresenhamLow(x1, y1, x0, y0)
		} else {
			r.bresenhamLow(x0, y0, x1, y1)
		}
	} else {
		if y0 > y1 {
			r.bresenhamHigh(x1, y1, x0, y0)
		} else {
			r.bresenhamHigh(x0, y0, x1, y1)
		}
	}
}

func ref[T any](v T) *T {
	return &v
}
