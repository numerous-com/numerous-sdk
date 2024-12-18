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
	plotRune           rune = '█'
	boxVertical        rune = '│'
	boxDownLeft        rune = '┐'
	boxHorizontal      rune = '─'
	boxUpLeftRight     rune = '┴'
	boxUpLeftRightDown rune = '┼'

	defaultPlotLineWidth int = 80
	maxPlotLineWidth     int = 120

	plotShortenedLabelDotsCount int = 2
	plotXLabelMinLen            int = 3
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

// Display the plots data with each output line prefixed with `prefix`, and taking up
// `height` lines.
func (p *Plot) Display(prefix string, height int) {
	cols, _, err := p.term.GetSize()
	if err != nil {
		cols = defaultPlotLineWidth
	}

	if cols > maxPlotLineWidth {
		cols = maxPlotLineWidth
	}

	axes := p.newAxes(cols, height)
	rasterCols := cols - len(prefix) - max(len(axes.yLabelMin), len(axes.yLabelMax)) - 1
	normalized := p.data.Normalize(float64(rasterCols-1), float64(height-1))
	r := newPlotRaster(rasterCols, height)

	var prev *timeseries.NormalizedPoint
	for _, cur := range normalized {
		if prev == nil {
			prev = ref(cur)
			continue
		}

		r.BresenhamLine(int(prev.X), int(prev.Y), int(cur.X), int(cur.Y))

		prev = ref(cur)
	}

	w := p.term.Writer()
	fmt.Fprint(w, p.render(r, axes, prefix, height))
}

// Render the plotted raster, and the axes with the given prefix, returning it as a
// string.
func (p *Plot) render(raster *plotRaster, axes plotAxes, prefix string, height int) string {
	s := ""
	for y := height - 1; y >= 0; y-- {
		ln := prefix + axes.renderYAxisAtHeight(y)
		ln += raster.renderRow(y)
		s += ln + "\n"
	}
	s += axes.renderXAxis(prefix) + "\n"

	return s
}

func (p *Plot) newAxes(width, height int) plotAxes {
	axes := plotAxes{
		xLabelMin: p.data.MinTimestamp().Format(time.RFC3339),
		xLabelMax: p.data.MaxTimestamp().Format(time.RFC3339),
		width:     width,
		height:    height,
	}

	if p.data.MinValue() == p.data.MaxValue() {
		axes.yLabelMin = fmt.Sprintf("%.2f", p.data.MinValue()-1.0)
		axes.yLabelMax = fmt.Sprintf("%.2f", p.data.MaxValue()+1.0)
	} else {
		axes.yLabelMin = fmt.Sprintf("%.2f", p.data.MinValue())
		axes.yLabelMax = fmt.Sprintf("%.2f", p.data.MaxValue())
	}

	return axes
}

// contains the information needed to render the axes of a plot
type plotAxes struct {
	height    int
	width     int
	yLabelMin string
	yLabelMax string
	xLabelMin string
	xLabelMax string
}

// Renders a single line of the (vertical) y-axis of a plot, and return it as a
// string.
// For the top line it will render the y-axis maximum value.
func (pa *plotAxes) renderYAxisAtHeight(y int) string {
	maxLabelLen := max(len(pa.yLabelMax), len(pa.yLabelMin))

	if y == pa.height-1 {
		return strings.Repeat(" ", maxLabelLen-len(pa.yLabelMax)) + pa.yLabelMax + string(boxDownLeft)
	}

	return strings.Repeat(" ", maxLabelLen) + string(boxVertical)
}

// Renders the (horizontal) x-axis of a plot and returns it as a string.
// It renders two lines - one containing the minimum y-value as well as the
// actual box-drawn line, and another containing the x-axis labels.
func (pa *plotAxes) renderXAxis(prefix string) string {
	maxYLabelLen := max(len(pa.yLabelMax), len(pa.yLabelMin))

	hz := prefix + strings.Repeat(" ", maxYLabelLen-len(pa.yLabelMin)) + pa.yLabelMin + string(boxUpLeftRightDown)
	hzRemaining := pa.width - utf8.RuneCountInString(hz)
	if hzRemaining > 1 {
		hz += strings.Repeat(string(boxHorizontal), hzRemaining-1) + string(boxDownLeft)
	}

	xLabelLn := prefix + strings.Repeat(" ", maxYLabelLen)
	xLabelSpacing := pa.width - len(xLabelLn) - len(pa.xLabelMin) - len(pa.xLabelMax)
	switch {
	case xLabelSpacing > 0:
		xLabelLn += pa.xLabelMin
		xLabelLn += strings.Repeat(" ", xLabelSpacing)
		xLabelLn += pa.xLabelMax

		return hz + "\n" + xLabelLn
	case pa.width-len(xLabelLn)-len(prefix) > 1+2*plotShortenedLabelDotsCount+2*plotXLabelMinLen:
		// enough space to print part of the x-axis label:
		//  * 1 empty space
		//  * dots on both sides
		//  * a minimum x-label length
		remainingSpace := pa.width - len(xLabelLn) - 2*plotShortenedLabelDotsCount - 1
		minXLabelCutLen := remainingSpace / 2 // nolint:mnd
		maxXLabelCutLen := remainingSpace / 2 // nolint:mnd

		xLabelLn += pa.xLabelMin[0:minXLabelCutLen] +
			strings.Repeat(".", plotShortenedLabelDotsCount) +
			" " +
			strings.Repeat(".", plotShortenedLabelDotsCount) +
			pa.xLabelMax[len(pa.xLabelMax)-maxXLabelCutLen:]

		return hz + "\n" + xLabelLn
	default:
		// if there is no space for x-label, draw x-axis line differently
		hz = prefix + strings.Repeat(" ", maxYLabelLen-len(pa.yLabelMin)) + pa.yLabelMin + string(boxUpLeftRight)
		hzRemaining := pa.width - utf8.RuneCountInString(hz)
		if hzRemaining > 1 {
			hz += strings.Repeat(string(boxHorizontal), hzRemaining)
		}

		return hz
	}
}

// A raster for plotting graphs onto.
// Has related methods for drawing lines onto the raster, and for rendering the
// raster to strings.
type plotRaster struct {
	data   []rune
	width  int
	height int
}

func newPlotRaster(width, height int) *plotRaster {
	return &plotRaster{
		data:   make([]rune, width*height),
		width:  width,
		height: height,
	}
}

// Render a single row (or line) of the raster, returning it as a string.
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

// Plot a data point, marking the specified cell and all cells below it with the
// specified rune.
func (r *plotRaster) plot(x, y int, val rune) {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return
	}

	for yPlot := y; yPlot >= 0; yPlot-- {
		i := yPlot*r.width + x
		r.data[i] = val
	}
}

// Draw a line onto the raster
// See: https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
func (r *plotRaster) BresenhamLine(x0, y0, x1, y1 int) {
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

func ref[T any](v T) *T {
	return &v
}
