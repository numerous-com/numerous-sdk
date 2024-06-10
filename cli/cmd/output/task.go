package output

import (
	"fmt"
)

const taskLineLength = 40

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
	fmt.Println(ansiReset+ansiFaint+prefix+ansiReset, line)
	t.lineAdded = true
}

func (t *Task) Done() {
	ln := t.line(checkmark)
	if !t.lineAdded {
		fmt.Print("\r")
	}
	fmt.Println(ln + ansiGreen + "OK" + ansiReset)
}

func (t Task) Error() {
	ln := t.line(errorcross)
	if !t.lineAdded {
		fmt.Print("\r")
	}
	fmt.Println(ln + ansiRed + "Error" + ansiReset)
}

func StartTask(msg string) *Task {
	t := Task{msg: msg, length: taskLineLength}
	t.start()

	return &t
}
