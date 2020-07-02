package main

import (
	"fmt"
	"testing"
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/src/di_messages/entry"
	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/utils"

	"github.com/golang/protobuf/proto"
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

			updatedAtDate := time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC)
			uptdateAt := utils.TimeToProtoTime(updatedAtDate)

			updateEntryMessage := &entryMessage.InfoEntryUpdated{
				Payload: &entryMessage.InfoEntryUpdated_Payload{
					Id:        "123",
					Text:      "Some updated entry",
					CreatorId: "1",
					UpdatedAt: &uptdateAt,
				},
			}

			updateEntryMessageRequest, err := proto.Marshal(updateEntryMessage)
			if err != nil {
				panic(err)
			}

			_, err = nc.Request("info.entry.updated", updateEntryMessageRequest, 2*time.Second)
			// TODO: Send more including ones on the same date

			if err != nil {
				panic(err)
			}

			startTime := time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC)
			start := utils.TimeToProtoTime(startTime)

			endTime := time.Date(2020, time.January, 1, 30, 2, 03, 04, time.UTC)
			end := utils.TimeToProtoTime(endTime)

			requestMessage := &insightsMessage.GetVelocityRequest{
				Payload: &insightsMessage.GetVelocityRequest_Payload{
					Start: &start,
					End:   &end,
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
