package main

import (
	"testing"
	"time"

	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	nats "github.com/nats-io/nats.go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIntegration(t *testing.T) {
	Convey("The system", t, func() {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		Convey("works", func() {
			startTime, err := time.Parse(time.RFC3339, "2020-01-01T10:02:03+04:00")
			if err != nil {
				panic(err)
			}

			start := &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			}

			endTime, err := time.Parse(time.RFC3339, "2020-01-30T10:02:03+04:00")
			if err != nil {
				panic(err)
			}

			end := &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			}

			requestMessage := &insightsMessage.GetVelocityRequest{
				Payload: &insightsMessage.GetVelocityRequest_Payload{
					Start: start,
					End:   end,
				},
			}
			request, err := proto.Marshal(requestMessage)
			if err != nil {
				panic(err)
			}

			responseMessage, err := nc.Request("insights.get.velocity", request, 5*time.Second)

			response := &insightsMessage.GetVelocityResponse{}
			err = proto.Unmarshal(responseMessage.Data, response)
			So(err, ShouldBeNil)
			So(len(response.Payload), ShouldEqual, 0)
		})
	})
}
