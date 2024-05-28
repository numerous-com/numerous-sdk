package output

import "fmt"

var (
	greenColorEscapeANSI = "\033[32m"
	resetColorEscapeANSI = "\033[0m"
	unicodeCheckmark     = "\u2713"
	greenCheckmark       = greenColorEscapeANSI + unicodeCheckmark + resetColorEscapeANSI
	unicodeHourglass     = "\u29D6"
)

func PrintTaskStarted(message string) {
	fmt.Print(unicodeHourglass + "  " + message)
}

func PrintTaskDone(message string) {
	fmt.Println("\r" + greenCheckmark + "  " + message + "Done")
}
