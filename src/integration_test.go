package main

import (
	"fmt"
	"testing"
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/src/di_messages/entry"
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
			nc.Request("insights.store.drop", []byte(""), 3*time.Second)

			updatedAtDate, err := time.Parse(time.RFC3339, "2020-01-01T10:02:03+04:00")
			if err != nil {
				panic(err)
			}
			updateEntryMessage := &entryMessage.InfoEntryUpdated{
				Payload: &entryMessage.InfoEntryUpdated_Payload{
					Id:        "123",
					Text:      "Some updated entry",
					CreatorId: "1",
					UpdatedAt: &timestamp.Timestamp{
						Seconds: int64(updatedAtDate.Unix()),
					},
				},
			}

			updateEntryMessageRequest, err := proto.Marshal(updateEntryMessage)
			if err != nil {
				panic(err)
			}

			_, err = nc.Request("info.entry.updated", updateEntryMessageRequest, 2*time.Second)

			if err != nil {
				panic(err)
			}

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
			fmt.Println(response.Payload)
			So(response.Payload[0].Score, ShouldEqual, 1)
		})
	})
}
