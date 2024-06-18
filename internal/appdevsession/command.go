package appdevsession

import (
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

type commandExecutor interface {
	// Create a command that runs the named command, with the provided arguments.
	Create(name string, args ...string) command
}

type command interface {
	// Kill the command
	Kill() error
	// Start the command
	Start() error
	// Set an environment variable for the command
	SetEnv(string, string)
	// Returns a reader for the commands standard output.
	StdoutPipe() (io.ReadCloser, error)
	// Returns a reader for the commands standard error.
	StderrPipe() (io.ReadCloser, error)
	// Runs the command, and returns its combined output (both standard output and standard error).
	Output() ([]byte, error)
}

type execCommandExecutor struct{}

func (e *execCommandExecutor) Create(name string, args ...string) command {
	slog.Debug("creating command", slog.String("name", name), slog.Any("args", args))
	cmd := &execCommand{cmd: *exec.Command(name, args...)}

	return cmd
}

type execCommand struct {
	cmd exec.Cmd
}

func (e *execCommand) Start() error {
	slog.Debug("starting command", slog.String("path", e.cmd.Path), slog.Any("args", e.cmd.Args))
	return e.cmd.Start()
}

func (e *execCommand) Kill() error {
	return e.cmd.Process.Kill()
}

func (e *execCommand) SetEnv(key string, value string) {
	e.cmd.Env = append(e.cmd.Env, fmt.Sprintf("%s=%s", key, value))
}

func (e *execCommand) StdoutPipe() (io.ReadCloser, error) {
	return e.cmd.StdoutPipe()
}

func (e *execCommand) StderrPipe() (io.ReadCloser, error) {
	return e.cmd.StderrPipe()
}

func (e *execCommand) Output() ([]byte, error) {
	return e.cmd.CombinedOutput()
}
