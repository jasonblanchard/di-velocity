package domain

import (
	"time"
)

// ExpandVelicityScores Accepts sparse list of daily velocities and returns list that includes zero'd velocity days between start and end
func ExpandVelicityScores(dailyVelocities DailyVelocities, start time.Time, end time.Time) DailyVelocities {
	var output DailyVelocities
	next := start

	var velocitiesByDate = make(map[time.Time]DailyVelocity)

	for i := 0; i < len(dailyVelocities); i++ {
		velocitiesByDate[dailyVelocities[i].Day] = dailyVelocities[i]
	}

	for next.Unix() <= end.Unix() {
		velocity, ok := velocitiesByDate[next]
		if ok != true {
			velocity = DailyVelocity{
				Day: next,
			}
		}
		output = append(output, velocity)
		next = next.AddDate(0, 0, 1)
	}

	return output
}
