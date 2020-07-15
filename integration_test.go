package main

import (
	"testing"
	"time"

	"github.com/jasonblanchard/di-velocity/internal/app"
	entryMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/entry"
	insightsMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/insights"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/nats.go"
	. "github.com/smartystreets/goconvey/convey"
)

func updateEntry(nc *nats.Conn, date time.Time) error {
	uptdateAt := app.TimeToProtoTime(date)

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

	err = nc.Publish("info.entry.updated", updateEntryMessageRequest)
	return err
}

func TestIntegration(t *testing.T) {
	Convey("The system", t, func() {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		Convey("works", func() {
			_, err := nc.Request("insights.store.drop", []byte(""), 3*time.Second)

			if err != nil {
				panic(err)
			}

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
			time.Sleep(3 * time.Second)

			if err != nil {
				panic(err)
			}

			startTime := time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC)
			start := app.TimeToProtoTime(startTime)

			endTime := time.Date(2020, time.January, 1, 30, 2, 03, 04, time.UTC)
			end := app.TimeToProtoTime(endTime)

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

			responseMessage, err := nc.Request("insights.get.velocity", request, 3*time.Second)
			if err != nil {
				panic(err)
			}

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
