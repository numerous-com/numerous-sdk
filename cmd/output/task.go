package output

import (
	"fmt"
	"io"
)

const taskLineLength = 48

type Task struct {
	msg       string
	length    int
	lineAdded bool
}

func (t *Task) line(icon string) string {
	ln := icon + " " + t.msg
	for d := len(t.msg); d < t.length; d++ {
		ln += "."
	}

	return ln
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

func StartTask(msg string) *Task {
	t := Task{msg: msg, length: taskLineLength}
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
