package app

import "time"

type StubClock struct{ now time.Time }

func (c *StubClock) Now() time.Time {
	return c.now
}
