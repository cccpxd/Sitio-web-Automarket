package clock

import "time"

// System is the production implementation of service.Clock.
type System struct{}

func New() System {
	return System{}
}

func (System) Now() time.Time {
	return time.Now().UTC()
}
