package domain

import (
	"time"

	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/utils"
)

// DailyVelocity represents a velocity score on a partiular day
type DailyVelocity struct {
	Day       time.Time
	Score     int32
	CreatorID string
}

// DailyVelocities collection of DailyVelocity entities
type DailyVelocities []DailyVelocity

// ToDtoPayload convert DailyVelocities to dto payload
func (dailyVelocities DailyVelocities) ToDtoPayload() []*insights.GetVelocityResponse_DailyVelocity {
	payload := make([]*insights.GetVelocityResponse_DailyVelocity, len(dailyVelocities))

	for i := 0; i < len(dailyVelocities); i++ {
		day := utils.TimeToProtoTime(dailyVelocities[i].Day)

		payload[i] = &insights.GetVelocityResponse_DailyVelocity{
			Day:   &day,
			Score: dailyVelocities[i].Score,
		}
	}

	return payload
}
