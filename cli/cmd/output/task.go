package output

import (
	"fmt"
)

const taskLineLength = 40

type Task struct {
	msg    string
	length int
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

func (t *Task) Done() {
	ln := t.line(checkmark)
	fmt.Println("\r" + ln + ansiGreen + "OK" + ansiReset)
}

func (t Task) Error() {
	ln := t.line(errorcross)
	fmt.Println("\r" + ln + ansiRed + "Error" + ansiReset)
}

func StartTask(msg string) *Task {
	t := Task{msg: msg, length: taskLineLength}
	t.start()

	return &t
}
