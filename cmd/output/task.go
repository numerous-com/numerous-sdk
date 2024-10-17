package output

import (
	"fmt"
	"io"
	"strings"
)

const (
	fallbackTaskLineWidth = 60
	maxTaskLineWidth      = 120
	minDots               = 3
)

type lineWidthFunc func() int

type Task struct {
	msg          string
	lineAdded    bool
	lineUpdating bool
	lineWidth    lineWidthFunc
	w            io.Writer
}

func (t *Task) line(icon string) string {
	dotCount, msg := t.trimMessage()

	line := icon + " " + msg + AnsiFaint + "..."
	line += strings.Repeat(".", dotCount)

	return line + AnsiReset
}

func (t *Task) trimMessage() (int, string) {
	w := t.lineWidth()

	errLineLen := 2 + len(t.msg) + minDots + len("Error") // +2 for icon and space
	lenDiff := w - errLineLen
	msg := t.msg
	dotCount := 0
	if lenDiff < 0 {
		msgEnd := len(t.msg) + lenDiff
		if msgEnd < 0 {
			msgEnd = 0
		}

		msg = t.msg[0:msgEnd]
	} else {
		dotCount = lenDiff
	}

	return dotCount, msg
}

func (t *Task) start() {
	ln := t.line(hourglassIcon)
	fmt.Fprint(t.w, ln)
}

func (t *Task) AddLine(prefix string, line string) {
	if !t.lineAdded || t.lineUpdating {
		fmt.Fprintln(t.w)
	}
	fmt.Fprintln(t.w, AnsiReset+AnsiFaint+prefix+AnsiReset, line)
	t.lineUpdating = false
	t.lineAdded = true
}

func (t *Task) UpdateLine(prefix string, line string) {
	if !t.lineAdded {
		fmt.Fprintln(t.w)
	}
	fmt.Fprint(t.w, "\r"+AnsiReset+AnsiFaint+prefix+AnsiReset+" "+line)
	t.lineUpdating = true
	t.lineAdded = true
}

func (t *Task) EndUpdateLine() {
	if t.lineUpdating {
		fmt.Fprintln(t.w)
		t.lineUpdating = false
	}
}

func (t *Task) Done() {
	t.terminate(checkmarkIcon, AnsiGreen+"OK"+AnsiReset)
}

func (t *Task) Error() {
	t.terminate(errorcross, AnsiRed+"Error"+AnsiReset)
}

func (t *Task) terminate(icon, status string) {
	ln := t.line(icon)
	if t.lineUpdating {
		fmt.Fprintln(t.w)
	}
	if !t.lineAdded {
		fmt.Fprint(t.w, "\r")
	}
	fmt.Fprintln(t.w, ln+status)
}

type terminal interface {
	IsTerminal() bool
	GetSize() (int, int, error)
	Writer() io.Writer
}

func terminalWidthFunc(t terminal) lineWidthFunc {
	// if we are not writing to a terminal (e.g. piping to a file) fall back to
	// fallback task line width
	if !t.IsTerminal() {
		return func() int { return fallbackTaskLineWidth }
	}

	// otherwise get terminal size
	return func() int {
		w, _, err := t.GetSize()
		switch {
		case err != nil:
			return fallbackTaskLineWidth
		case w < maxTaskLineWidth:
			return w
		default:
			return maxTaskLineWidth
		}
	}
}

func StartTask(msg string) *Task {
	return StartTaskWithTerminal(msg, newTermTerminal())
}

func StartTaskWithTerminal(msg string, t terminal) *Task {
	f := terminalWidthFunc(t)
	task := Task{msg: msg, lineWidth: f, w: t.Writer()}
	task.start()

	return &task
}

var _ io.Writer = &TaskLineWriter{}

type TaskLineWriter struct {
	task   *Task
	prefix string
}

func NewTaskLineWriter(t *Task, prefix string) *TaskLineWriter {
	return &TaskLineWriter{t, prefix}
}

func (tlw *TaskLineWriter) Write(buf []byte) (int, error) {
	tlw.task.AddLine(tlw.prefix, string(buf))
	return len(buf), nil
}
