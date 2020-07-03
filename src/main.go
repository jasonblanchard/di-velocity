package main

import (
	"database/sql"
	"flag"
	"fmt"
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
	"github.com/spf13/viper"

	_ "github.com/lib/pq"

	nats "github.com/nats-io/nats.go"
)

var natsQueue = "valocity"

func initConfig(path string) string {
	if path != "" {
		viper.SetConfigFile(path)
	}
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if err != nil {
		log.Info().Msg(err.Error())
		return ""
	}
	return viper.ConfigFileUsed()
}

func main() {
	config := flag.String("config", "", "Path to config file")
	flag.Parse()

	configFile := initConfig(*config)

	if viper.GetBool("pretty") == true {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetBool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if configFile != "" {
		log.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", viper.GetString("db_user"), viper.GetString("db_password"), viper.GetString("db_name"))
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

		incrementDailyCounterRequest := &insightsMessage.IncrementDailyCounter{
			Payload: &insightsMessage.IncrementDailyCounter_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(incrementDailyCounterRequest)

		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		nc.Publish("insights.increment.dailyCounter", request)

		if m.Reply != "" {
			nc.Publish(m.Reply, []byte(""))
		}
	})

	nc.QueueSubscribe("insights.increment.dailyCounter", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")

		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = op.IncrementDailyCounter(db, day, requestMessage.Payload.CreatorId)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}
	})

	// TODO: Enable in test mode only
	nc.QueueSubscribe("insights.store.drop", natsQueue, func(m *nats.Msg) {
		log.Info().
			Str("subject", m.Subject).
			Msg("received")

		err := op.DropDailyCounts(db)
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

		dailyCounts, err := op.GetDailyCounts(db, normalizedStart, normalizedEnd)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
			return
		}

		dailyVelocities := dailyCounts.ToVelocityScores()

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
