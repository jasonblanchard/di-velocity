package domain

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
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
		payload[i] = &insights.GetVelocityResponse_DailyVelocity{
			Day: &timestamp.Timestamp{
				Seconds: dailyVelocities[i].Day.Unix(),
			},
			Score: dailyVelocities[i].Score,
		}
	}

	return payload
}
