package report

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

const numerousReportURL = "https://numerous.com"

var ReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Opens numerous report and feedback page.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := openURL(numerousReportURL); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func openURL(url string) error {
	cmd, args, err := setCmdByOS()
	if err != nil {
		return err
	}
	fmt.Println("Opening the report page in your default browser.")
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

func setCmdByOS() (string, []string, error) {
	switch runtime.GOOS {
	case "windows":
		return "cmd", []string{"/c", "start"}, nil
	case "darwin":
		return "open", nil, nil
	case "freebsd", "openbsd", "netbsd":
		return "xdg-open", nil, nil
	case "linux":
		out, err := exec.Command("sh", "-c", "grep -i Microsoft /proc/version").Output()
		if err != nil {
			fmt.Println("Error:", err)
			return "", nil, err
		}
		if string(out) != "" {
			return "sensible-browser", nil, nil
		} else {
			return "xdg-open", nil, nil
		}
	default:
		return "", nil, errors.New("it wasn't possible to identify your OS")
	}
}
