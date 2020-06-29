package main

import (
	"flag"
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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	nats "github.com/nats-io/nats.go"
)

var natsQueue = "valocity"

func main() {
	pretty := flag.Bool("pretty", false, "Pretty print logs")
	debugLoglevel := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	if *pretty == true {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debugLoglevel {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg(">>> Starting <<<")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
		os.Exit(1)
	}

	nc.QueueSubscribe("insights.get.velocity", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			log.Error().
				Str("subject", m.Subject).
				Err(err).
				Msg("")

			return
			// TODO: Respond with error type
		}

		normalizedStart := utils.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := utils.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyVelocities, err := op.GetDailyVelocity(normalizedStart, normalizedEnd)
		if err != nil {
			log.Error().
				Str("subject", m.Subject).
				Err(err).
				Msg("")
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
			log.Error().
				Str("subject", m.Subject).
				Err(err).
				Msg("")
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

	log.Info().Msg("Ready to receive messages")
	runtime.Goexit()
}
