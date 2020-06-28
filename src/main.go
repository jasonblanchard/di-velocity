package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/op"
	"github.com/jasonblanchard/di-velocity/src/utils"

	nats "github.com/nats-io/nats.go"
)

func main() {
	fmt.Println(">>> Starting <<<")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	nc.Subscribe("insights.get.velocity", func(m *nats.Msg) {
		fmt.Println("receiving " + m.Subject)
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			panic(err)
		}

		normalizedStart := utils.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := utils.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyVelocities, err := op.GetDailyVelocity(normalizedStart, normalizedEnd)
		if err != nil {
			panic(err)
		}

		payload := make([]*insights.GetVelocityResponse_DailyVelocity, len(dailyVelocities))

		for i := 0; i < len(dailyVelocities); i++ {
			payload = append(payload, &insights.GetVelocityResponse_DailyVelocity{
				Day: &timestamp.Timestamp{
					Seconds: int64(dailyVelocities[i].Day.Second()),
				},
				Score: dailyVelocities[i].Score,
			})
		}

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: payload,
		}

		response, err := proto.Marshal(responseMessage)
		if err != nil {
			panic(err)
		}

		nc.Publish(m.Reply, response)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		// Wait for signal
		<-c
		nc.Drain()
		os.Exit(0)
	}()
	runtime.Goexit()
}
