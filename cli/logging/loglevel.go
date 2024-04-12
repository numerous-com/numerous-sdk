package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

var ErrInvalidLogLevel error = errors.New(`must be one of "debug", "info", "warning" or "error"`)

func (l *Level) String() string {
	return string(*l)
}

func (l *Level) Set(v string) error {
	v = strings.ToLower(v)
	switch v {
	case "debug", "info", "warning", "error":
		*l = Level(v)
		return nil
	default:
		return ErrInvalidLogLevel
	}
}

func (l *Level) Type() string {
	return "Log level"
}

func (l *Level) ToSlogLevel() slog.Level {
	switch strings.ToLower(l.String()) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic(fmt.Sprintf("unexpected log level \"%s\"", l.String()))
	}
}
