package output

import (
	"fmt"
	"math/rand"
)

var notifyCmdMovedBody = AnsiFaint + "A new set of commands related to apps in organizations have been promoted as \n" +
	"the default commands, and previous commands have been moved to the \"legacy\" namespace." + AnsiReset + "\n\n" +
	"See https://www.numerous.com/docs/cli#legacy-commands for more information."

func Notify(header, body string, args ...any) {
	body = addTrailingNewLine(body)
	fmt.Printf(AnsiCyanBold+header+AnsiReset+"\n"+body, args...)
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
	feedbackHeader = "Thank you for using Numerous!"
	feedbackURL    = AnsiBlue + "https://www.numerous.com/app/feedback" + AnsiReset
	feedbackBody   = `Numerous is still in development and as such, we are very open to hearing your
feedback. If you experience issues or have improvement suggestions, please
visit:`
)

func NotifyFeedbackMaybe() {
	probability := 0.1
	if rand.Float64() >= probability {
		return
	}

	Notify(feedbackHeader, feedbackBody+" "+feedbackURL)
	fmt.Println()
}
