package list

import (
	"errors"
	"strings"
)

type DisplayMode string

const (
	DisplayModeList  DisplayMode = "list"
	DisplayModeTable DisplayMode = "table"
)

var errInvalidDisplayMode error = errors.New(`must be one of "list", or "table"`)

func (l *DisplayMode) String() string {
	return string(*l)
}

func (l *DisplayMode) Set(v string) error {
	v = strings.ToLower(v)
	switch v {
	case "list", "table":
		*l = DisplayMode(v)
		return nil
	default:
		return errInvalidDisplayMode
	}
}

func (l *DisplayMode) Type() string {
	return "Display mode"
}
