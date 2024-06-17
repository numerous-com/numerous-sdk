package output

import "runtime"

const (
	AnsiRed      = "\033[31m"
	AnsiGreen    = "\033[32m"
	AnsiYellow   = "\033[33m"
	AnsiReset    = "\033[0m"
	AnsiFaint    = "\033[2m"
	AnsiCyanBold = "\033[1;36m"
)

func highlight(value string) string {
	if runtime.GOOS == "windows" {
		return "\"" + value + "\""
	} else {
		return AnsiCyanBold + value + AnsiReset
	}
}
