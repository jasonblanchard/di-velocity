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

// ToVelocityScores calculate velocity scores by daily counts
func (counts DailyCounts) ToVelocityScores() DailyVelocities {
	dailyVelocities := DailyVelocities{}

	for i := 0; i < len(counts); i++ {
		dailyVelocity := DailyVelocity{
			Day:   counts[i].Day.UTC(),
			Score: CountToScore(counts[i].Count), // TODO: Convert to score
		}
		dailyVelocities = append(dailyVelocities, dailyVelocity)
	}

	return dailyVelocities
}
