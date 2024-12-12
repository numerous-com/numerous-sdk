package app

import "time"

type AppDeployLogEntry struct {
	Timestamp time.Time
	Text      string
}
