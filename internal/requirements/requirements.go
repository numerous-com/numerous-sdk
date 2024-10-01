package requirements

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

type requirementsTxt struct {
	encoder *encoding.Encoder
	lines   []string
	crlf    bool
}

func Read(r io.Reader) (*requirementsTxt, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	encoding := unicode.UTF8
	for _, bom := range boms {
		if bom.Detect(data) {
			encoding = bom.Encoding()
			break
		}
	}

	decoder := encoding.NewDecoder()
	data, err = decoder.Bytes(data)
	if err != nil {
		return nil, err
	}

	lines, crln, err := readLines(data)
	if err != nil {
		return nil, err
	}

	return &requirementsTxt{
		encoder: encoding.NewEncoder(),
		lines:   lines,
		crlf:    crln,
	}, nil
}

func readLines(data []byte) (lines []string, crln bool, err error) {
	lines = []string{}
	crln = false
	returnedFinalNonTerminatedLine := false
	returnedLastEmptyLine := false

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			if !returnedFinalNonTerminatedLine && !returnedLastEmptyLine {
				returnedLastEmptyLine = true
				return 0, []byte{}, nil
			}

			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			token, lineHadCR := dropCR(data[0:i])
			crln = crln || lineHadCR

			return i + 1, token, nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			token, lineCR := dropCR(data)
			crln = crln || lineCR
			returnedFinalNonTerminatedLine = true

			return len(data), token, nil
		}
		// Request more data.
		return 0, nil, nil
	}

	s := bufio.NewScanner(bytes.NewBuffer(data))
	s.Split(split)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, false, err
		}

		lines = append(lines, s.Text())
	}

	return lines, crln, nil
}

// dropCR drops a terminal \r from the data, and return true.
func dropCR(data []byte) ([]byte, bool) {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1], true
	}

	return data, false
}

func (r *requirementsTxt) Write(w io.Writer) error {
	ew := r.encoder.Writer(w)

	lineEnding := []byte("\n")
	if r.crlf {
		lineEnding = []byte("\r\n")
	}

	for i, l := range r.lines {
		if i > 0 {
			if _, err := ew.Write(lineEnding); err != nil {
				return err
			}
		}

		if _, err := ew.Write([]byte(l)); err != nil {
			return err
		}
	}

	return nil
}

func (r *requirementsTxt) Add(added string) {
	for _, l := range r.lines {
		if strings.Contains(l, added) {
			return
		}
	}

	lastLine := r.lines[len(r.lines)-1]
	if lastLine == "" {
		// if the last line is empty, insert above the empty line
		r.lines[len(r.lines)-1] = added
		r.lines = append(r.lines, "")
	} else {
		r.lines = append(r.lines, added)
		r.lines = append(r.lines, "")
	}
}
