package op

import "time"

// DailyVelocity represents a velocity score on a partiular day
type DailyVelocity struct {
	Day   time.Time
	Score int32
}

// GetDailyVelocity returns velocity score for each day between start and end (inclusive)
func GetDailyVelocity(start time.Time, end time.Time) ([]DailyVelocity, error) {

	return make([]DailyVelocity, 0), nil
}
