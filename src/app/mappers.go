package app

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/domain"
)

// VelocitiesToProtoPayload converts velocity domain objects to protobuf payload
func VelocitiesToProtoPayload(dailyVelocities domain.DailyVelocities) []*insights.GetVelocityResponse_DailyVelocity {
	payload := make([]*insights.GetVelocityResponse_DailyVelocity, len(dailyVelocities))

	for i := 0; i < len(dailyVelocities); i++ {
		day := TimeToProtoTime(dailyVelocities[i].Day)

		payload[i] = &insights.GetVelocityResponse_DailyVelocity{
			Day:   &day,
			Score: dailyVelocities[i].Score,
		}
	}

	return payload
}

// TimeToProtoTime converts Go time objec to protobuf time object
func TimeToProtoTime(time time.Time) timestamp.Timestamp {
	return timestamp.Timestamp{
		Seconds: time.Unix(),
	}
}
