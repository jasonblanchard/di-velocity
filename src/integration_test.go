package main

import (
	"testing"
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/src/di_messages/entry"
	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/utils"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/nats.go"
	. "github.com/smartystreets/goconvey/convey"
)

func updateEntry(nc *nats.Conn, date time.Time) error {
	uptdateAt := utils.TimeToProtoTime(date)

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
		return err
	}

	_, err = nc.Request("info.entry.updated", updateEntryMessageRequest, 2*time.Second)
	return err
}

func TestIntegration(t *testing.T) {
	Convey("The system", t, func() {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		Convey("works", func() {
			nc.Request("insights.store.drop", []byte(""), 3*time.Second)

			updateEntry(nc, time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC))

			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2020, time.January, 2, 12, 2, 03, 04, time.UTC))

			updateEntry(nc, time.Date(2019, time.January, 2, 12, 2, 03, 04, time.UTC))
			updateEntry(nc, time.Date(2021, time.January, 2, 12, 2, 03, 04, time.UTC))

			// Let everything resolve
			time.Sleep(2 * time.Second)

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
			So(len(response.Payload), ShouldEqual, 2)
			So(response.Payload[0].Day.Seconds, ShouldEqual, 1577836800)
			So(response.Payload[0].Score, ShouldEqual, 1)
			So(response.Payload[1].Day.Seconds, ShouldEqual, 1577923200)
			So(response.Payload[1].Score, ShouldEqual, 2)
		})
	})
}
