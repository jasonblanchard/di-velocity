package domain

import (
	"time"
)

// DailyCount count of edits on a given day
type DailyCount struct {
	Day       time.Time
	Count     int32
	CreatorID string
}

// DailyCounts collection of DailyCount entities
type DailyCounts []DailyCount

// ToVelocities calculate velocity scores by daily counts
func (counts DailyCounts) ToVelocities() DailyVelocities {
	dailyVelocities := DailyVelocities{}

	for i := 0; i < len(counts); i++ {
		dailyVelocity := DailyVelocity{
			Day:   counts[i].Day,
			Score: CountToScore(counts[i].Count),
		}
		dailyVelocities = append(dailyVelocities, dailyVelocity)
	}

	return dailyVelocities
}
