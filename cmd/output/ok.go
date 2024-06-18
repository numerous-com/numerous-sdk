package output

import "fmt"

// Prints message as a line prefixed with a green checkmark. Additional
// arguments are used for formatting.
func PrintlnOK(message string, args ...any) {
	fmt.Printf(AnsiGreen+checkmark+AnsiReset+" "+message+"\n", args...)
}
