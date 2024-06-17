package output

import "runtime"

const (
	hourglass    = "\u29D6"
	errorcross   = AnsiRed + "\u2715" + AnsiReset
	checkmark    = AnsiGreen + "\u2713" + AnsiReset
	raiseHand    = "\U0000270B"
	shootingStar = "\U0001F320"
)

func symbol(value string) string {
	if runtime.GOOS == "windows" {
		return ""
	} else {
		return value
	}
}
