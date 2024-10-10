package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	fallbackTaskLineWidth = 60
	maxTaskLineWidth      = 120
	minDots               = 3
)

type lineWidthFunc func() int

type Task struct {
	msg       string
	lineAdded bool
	lineWidth lineWidthFunc
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
	ln := t.line(hourglass)
	fmt.Print(ln)
}

func (t *Task) AddLine(prefix string, line string) {
	if !t.lineAdded {
		fmt.Println()
	}
	fmt.Println(AnsiReset+AnsiFaint+prefix+AnsiReset, line)
	t.lineAdded = true
}

func (t *Task) Done() {
	ln := t.line(checkmark)
	if !t.lineAdded {
		fmt.Print("\r")
	}
	fmt.Println(ln + AnsiGreen + "OK" + AnsiReset)
}

func (t *Task) Error() {
	ln := t.line(errorcross)
	if !t.lineAdded {
		fmt.Print("\r")
	}
	fmt.Println(ln + AnsiRed + "Error" + AnsiReset)
}

func terminalWidthFunc() lineWidthFunc {
	stdout := int(os.Stdout.Fd())

	// if we are not writing to a terminal (e.g. piping to a file) fall back to
	// fallback task line width
	if !term.IsTerminal(stdout) {
		return func() int { return fallbackTaskLineWidth }
	}

	// otherwise get terminal size
	return func() int {
		w, _, err := term.GetSize(stdout)
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
	f := terminalWidthFunc()
	t := Task{msg: msg, lineWidth: f}
	t.start()

	return &t
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
