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

var (
	feedbackHeader = AnsiCyanBold + "Thank you for using Numerous!" + AnsiReset
	feedbackURL    = AnsiBlue + "https://www.numerous.com/app/feedback" + AnsiReset
	feedbackBody   = `Numerous is under active development, and your feedback is very welcome. If you
experience issues, or have improvement suggestions, please visit:`
)

func NotifyFeedback() {
	fmt.Println(feedbackHeader)
	fmt.Println(feedbackBody)
	fmt.Println("    " + feedbackURL)
}
