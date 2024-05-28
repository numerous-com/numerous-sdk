package output

import (
	"fmt"
)

var (
	greenColorEscapeANSI = "\033[32m"
	resetColorEscapeANSI = "\033[0m"
	unicodeCheckmark     = "\u2713"
	greenCheckmark       = greenColorEscapeANSI + unicodeCheckmark + resetColorEscapeANSI
	unicodeHourglass     = "\u29D6"
	taskLineLength       = 40
)

type Task struct {
	msg    string
	length int
}

func (t *Task) line(icon string) string {
	ln := " " + icon + " " + t.msg
	for d := len(t.msg); d < t.length; d++ {
		ln += "."
	}

	return ln
}

func (t *Task) start() {
	ln := t.line(unicodeHourglass)
	fmt.Print(ln)
}

func (t *Task) Done() {
	ln := t.line(greenCheckmark)
	fmt.Println("\r" + ln + "OK")
}

func StartTask(msg string) *Task {
	t := Task{msg: msg, length: taskLineLength}
	t.start()

	return &t
}
