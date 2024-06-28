package report

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
)

const numerousReportURL = "https://numerous.com"

var (
	ErrWSLCheck        = errors.New("error when checking windows subsystem for linux")
	ErrOSCheck         = errors.New("error identifying the current OS")
	ErrGrepBashCommand = errors.New("error in the grep bash command execution")
)

var ReportCmd = &cobra.Command{
	Use:     "report",
	Short:   "Opens Numerous report and feedback page",
	Args:    cobra.NoArgs,
	GroupID: group.AdditionalCommandsGroupID,
	Run: func(cmd *cobra.Command, args []string) {
		if err := openURL(numerousReportURL, runtime.GOOS, Exec{}); err != nil {
			output.PrintErrorDetails("Error occurred opening feedback and report URL", err)
		}
	},
}

type CommandExecutor interface {
	Output(command string, args ...string) ([]byte, error)
	Start(command string, args ...string) error
}

type Exec struct{}

func (oe Exec) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

func (oe Exec) Start(command string, args ...string) error {
	return exec.Command(command, args...).Start()
}

func openURL(url string, os string, exec CommandExecutor) error {
	cmd, args, err := cmdByOS(os, exec, osOrWSL)
	if err != nil {
		output.PrintErrorDetails("An error occurred when opening the numerous report page.", err)
		return err
	}
	args = append(args, url)

	return exec.Start(cmd, args...)
}

func cmdByOS(os string, exec CommandExecutor, osOrWSL func(exec CommandExecutor) (string, error)) (string, []string, error) {
	switch os {
	case "windows":
		return "cmd", []string{"/c", "start"}, nil
	case "darwin":
		return "open", nil, nil
	case "freebsd", "openbsd", "netbsd":
		return "xdg-open", nil, nil
	case "linux":
		check, err := osOrWSL(exec)
		if err != nil {
			return "", nil, ErrWSLCheck
		}
		if check == "wsl" {
			return "sensible-browser", nil, nil
		} else {
			return "xdg-open", nil, nil
		}
	default:
		return "", nil, ErrOSCheck
	}
}

func osOrWSL(exec CommandExecutor) (string, error) {
	out, err := exec.Output("sh", "-c", "grep -i Windows /proc/version")
	if err != nil && err.Error() != "exit status 1" {
		return "", ErrGrepBashCommand
	}
	if len(out) > 0 {
		return "wsl", nil
	} else {
		return "os", nil
	}
}
