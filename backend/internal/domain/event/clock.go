package event

import "time"

// Clock is a port for obtaining the current time, enabling testable time control.
type Clock interface {
	Now() time.Time
}

// RealClock is the production implementation using time.Now.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
