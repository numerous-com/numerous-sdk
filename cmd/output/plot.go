package output

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"numerous.com/cli/internal/timeseries"
)

const (
	plotRune       rune = '#'
	boxVertical    rune = '│'
	boxDownLeft    rune = '┐'
	boxHorizontal  rune = '─'
	boxUpLeftRight rune = '┴'

	defaultPlotLineWidth int = 80
	maxPlotLineWidth     int = 120
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

	axes := plotAxes{
		xLabelMin: p.data.MinTimestamp().Format(time.RFC3339),
		xLabelMax: p.data.MaxTimestamp().Format(time.RFC3339),
	}

	if p.data.MinValue() == p.data.MaxValue() {
		axes.yLabelMin = fmt.Sprintf("%.2f", p.data.MinValue()-1.0)
		axes.yLabelMax = fmt.Sprintf("%.2f", p.data.MaxValue()+1.0)
	} else {
		axes.yLabelMin = fmt.Sprintf("%.2f", p.data.MinValue())
		axes.yLabelMax = fmt.Sprintf("%.2f", p.data.MaxValue())
	}

	rasterCols := cols - len(prefix) - max(len(axes.yLabelMin), len(axes.yLabelMax)) - 1
	normalized := p.data.Normalize(float64(rasterCols-1), float64(height-1))
	r := newRaster(rasterCols, height)

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
	fmt.Fprint(w, p.render(r, prefix, cols, height, axes))
}

func (p *Plot) render(raster *plotRaster, prefix string, width int, height int, axes plotAxes) string {
	s := ""
	for y := height - 1; y >= 0; y-- {
		ln := prefix + p.renderLeftAxisAtHeight(y, height, axes)
		ln += raster.renderRow(y)
		s += ln + "\n"
	}
	s += p.renderHorizontalAxis(axes, prefix, width) + "\n"

	return s
}

func (p *Plot) renderHorizontalAxis(axes plotAxes, prefix string, width int) string {
	maxYLabelLen := max(len(axes.yLabelMax), len(axes.yLabelMin))

	hz := strings.Repeat(" ", maxYLabelLen-len(axes.yLabelMin)) + axes.yLabelMin + string(boxUpLeftRight)
	hzRemaining := width - utf8.RuneCountInString(hz) - len(prefix)
	if hzRemaining > 0 {
		hz += strings.Repeat(string(boxHorizontal), hzRemaining)
	}

	xLabelLn := strings.Repeat(" ", maxYLabelLen)
	xLabelSpacing := width - len(xLabelLn) - len(axes.xLabelMin) - len(axes.xLabelMax) - len(prefix)
	if xLabelSpacing > 0 {
		xLabelLn += axes.xLabelMin
		xLabelLn += strings.Repeat(" ", xLabelSpacing)
		xLabelLn += axes.xLabelMax
	}

	return prefix + hz + "\n" + prefix + xLabelLn
}

func (p *Plot) renderLeftAxisAtHeight(y int, height int, axes plotAxes) string {
	maxLabelLen := max(len(axes.yLabelMax), len(axes.yLabelMin))

	if y == height-1 {
		return strings.Repeat(" ", maxLabelLen-len(axes.yLabelMax)) + axes.yLabelMax + string(boxDownLeft)
	}

	return strings.Repeat(" ", maxLabelLen) + string(boxVertical)
}

type plotAxes struct {
	yLabelMin string
	yLabelMax string
	xLabelMin string
	xLabelMax string
}

func newRaster(width, height int) *plotRaster {
	return &plotRaster{
		data:   make([]rune, width*height),
		width:  width,
		height: height,
	}
}

type plotRaster struct {
	data   []rune
	width  int
	height int
}

func (r *plotRaster) renderRow(y int) string {
	ln := ""
	for x := 0; x < r.width; x++ {
		pt := r.data[r.width*y+x]
		if pt == 0 {
			ln += " "
		} else {
			ln += string(pt)
		}
	}

	return ln
}

// plot a data point, marking the specified cell and all cells below it
func (r *plotRaster) plot(x, y int, val rune) {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return
	}

	for yPlot := y; yPlot >= 0; yPlot-- {
		i := yPlot*r.width + x
		r.data[i] = val
	}
}

func (r *plotRaster) bresenhamLow(x0, y0, x1, y1 int) {
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

func (r *plotRaster) bresenhamHigh(x0, y0, x1, y1 int) {
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

func (r *plotRaster) bresenhamLine(x0, y0, x1, y1 int) {
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
