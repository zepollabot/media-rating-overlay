package model

import "time"

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}
