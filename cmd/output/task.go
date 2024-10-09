package output

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	taskLineLength = 80
	minDots        = 3
)

type lineWidthFunc func() int

type Task struct {
	msg       string
	lineAdded bool
	lineWidth lineWidthFunc
}

func (t *Task) line(icon string) string {
	lenDiff, msg := t.trimMessage()

	line := icon + " " + msg + AnsiFaint + "..."
	for i := 0; i < lenDiff; i++ {
		line += "."
	}

	return line + AnsiReset
}

func (t *Task) trimMessage() (int, string) {
	w := t.lineWidth()

	errLineLen := 2 + len(t.msg) + minDots + len("Error") // +2 for icon and space
	lenDiff := w - errLineLen
	msg := t.msg
	if lenDiff < 0 {
		msgEnd := len(t.msg) + lenDiff
		if msgEnd < 0 {
			msgEnd = 0
		}

		msg = t.msg[0:msgEnd]
	}

	return lenDiff, msg
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
	if !term.IsTerminal(stdout) {
		return func() int { return taskLineLength }
	}

	return func() int {
		w, _, err := term.GetSize(stdout)
		if err != nil {
			return taskLineLength
		} else {
			return w
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
