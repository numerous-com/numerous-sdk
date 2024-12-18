package output

import (
	"io"
	"os"

	"golang.org/x/term"
)

type terminal interface {
	IsTerminal() bool
	GetSize() (width int, height int, err error)
	Writer() io.Writer
}

type termTerminal struct {
	fd int
	w  io.Writer
}

func newTermTerminal() *termTerminal {
	f := os.Stdout
	return &termTerminal{
		fd: int(f.Fd()),
		w:  f,
	}
}

func (t *termTerminal) Writer() io.Writer {
	return t.w
}

func (t *termTerminal) IsTerminal() bool {
	return term.IsTerminal(t.fd)
}

func (t *termTerminal) GetSize() (int, int, error) {
	return term.GetSize(t.fd)
}
