package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	entryMessage "github.com/jasonblanchard/di-velocity/src/di_messages/entry"
	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/op"
	"github.com/jasonblanchard/di-velocity/src/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"

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

	// connStr := "postgres://di:di@localhost:5432/di_velocity?sslmode=disable"
	connStr := "user=di password=di dbname=di_velocity sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
		os.Exit(1)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info().Msg(">>> Starting <<<")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
		os.Exit(1)
	}

	nc.QueueSubscribe("info.entry.updated", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")

		entryUpdatedMessage := &entryMessage.InfoEntryUpdated{}
		err := proto.Unmarshal(m.Data, entryUpdatedMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		normalizedDay := utils.NormalizeTime(time.Unix(entryUpdatedMessage.Payload.UpdatedAt.Seconds, 0))
		day := utils.TimeToProtoTime(normalizedDay)

		registerVelocityEventRequest := &insightsMessage.RegisterVelocityEvent{
			Payload: &insightsMessage.RegisterVelocityEvent_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(registerVelocityEventRequest)

		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		nc.Publish("insights.register.velocity", request)

		if m.Reply != "" {
			nc.Publish(m.Reply, []byte(""))
		}
	})

	nc.QueueSubscribe("insights.register.velocity", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")

		requestMessage := &insights.RegisterVelocityEvent{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = op.RegisterVelocity(db, day, requestMessage.Payload.CreatorId)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}
	})

	// TODO: Enable in test mode only
	nc.QueueSubscribe("insights.store.drop", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")

		err := op.DropDailyVelocities(db)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		log.Info().
			Str("subject", m.Subject).
			Msg("complete")
		nc.Publish(m.Reply, []byte(""))
	})

	nc.QueueSubscribe("insights.get.velocity", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
			return
			// TODO: Respond with error type
		}

		normalizedStart := utils.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := utils.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyVelocities, err := op.GetDailyVelocity(db, normalizedStart, normalizedEnd)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
			return
		}

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: dailyVelocities.ToDtoPayload(),
		}

		response, err := proto.Marshal(responseMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		nc.Publish(m.Reply, response)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		// Wait for signal
		<-c
		db.Close()
		nc.Drain()
		os.Exit(0)
	}()

	log.Info().Msg("Ready to receive messages")
	runtime.Goexit()
}
