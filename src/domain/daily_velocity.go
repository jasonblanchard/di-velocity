package domain

import (
	"time"
)

// DailyVelocity represents a velocity score on a partiular day
type DailyVelocity struct {
	Day       time.Time
	Score     int32
	CreatorID string
}

// DailyVelocities collection of DailyVelocity entities
type DailyVelocities []DailyVelocity
