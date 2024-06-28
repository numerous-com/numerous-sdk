package output

import "fmt"

var notifyCmdMovedBody = AnsiFaint + "A new set of commands related to apps in organizations have been promoted as \n" +
	"the default commands, and previous commands have been moved to the \"legacy\" namespace." + AnsiReset + "\n\n" +
	"See https://www.numerous.com/docs/cli#legacy-commands for more information."

func Notify(header, body string, args ...any) {
	fmt.Printf(AnsiCyanBold+header+AnsiReset+"\n"+body+"\n", args...)
}

func NotifyCmdMoved(cmd string, newCmd string) {
	Notify(
		"Please note that the command \"%s\" has moved to \"%s\"",
		notifyCmdMovedBody,
		cmd, newCmd,
	)
}

func NotifyCmdChanged(cmd string, newCmd string) {
	Notify(
		"Please note that the command \"%s\" has changed, and the original command is moved to \"%s\"",
		notifyCmdMovedBody,
		cmd, newCmd,
	)
}
