package output

import "runtime"

const (
	hourglassIcon = "\u29D6"
	errorcross    = AnsiRed + "\u2715" + AnsiReset
	checkmarkIcon = AnsiGreen + "\u2713" + AnsiReset
	raiseHand     = "\U0000270B"
	shootingStar  = "\U0001F320"
)

func symbol(value string) string {
	if runtime.GOOS == "windows" {
		return ""
	} else {
		return value
	}
}
