package main

import (
	"fmt"
	"os"
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/di_messages/entry"
	errorMessage "github.com/jasonblanchard/di-velocity/di_messages/error"
	"github.com/jasonblanchard/di-velocity/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/domain"
	"github.com/jasonblanchard/di-velocity/mappers"
	"github.com/jasonblanchard/di-velocity/repository"
	"github.com/jasonblanchard/natsby"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

// Container holds all service dependencies
type Container struct {
	NATSConnection *nats.Conn
	Logger         *zerolog.Logger
	NATSQueue      string
	TestMode       bool
	Repository     repository.T
}

// ContainerInput arg for Service
type ContainerInput struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDbName   string
	PostgresHost     string
	NatsURL          string
	Debug            bool
	Pretty           bool
	TestMode         bool
	NATSQueue        string
}

// NewContainer initializes service dependencies
func NewContainer(input *ContainerInput) (*Container, error) {
	container := &Container{
		NATSQueue: input.NATSQueue,
		TestMode:  input.TestMode,
	}

	repository, err := repository.NewPostgres(input.PostgresUser, input.PostgresPassword, input.PostgresDbName, input.PostgresHost)
	if err != nil {
		return container, errors.Wrap(err, "Repository initialization failed")
	}

	container.Repository = repository

	nc, err := initializeNATS(input.NatsURL)
	if err != nil {
		return container, errors.Wrap(err, "NATS initialization failed")
	}

	container.NATSConnection = nc

	logger := initializeLogger(input.Debug, input.Pretty)
	container.Logger = logger

	return container, nil
}

func initializeLogger(debug, pretty bool) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if pretty == true {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug == true {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.DurationFieldUnit = time.Second
	return &logger
}

// InitializeNATS sets up NATS connection
func initializeNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// InitExternalConfig pull in configuration from environment or file
func InitExternalConfig(path string) string {
	if path != "" {
		viper.SetConfigFile(path)
	}
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if err != nil {
		return ""
	}
	return viper.ConfigFileUsed()
}

// SubscribeHandlers configures all the handlers
func (container *Container) SubscribeHandlers(e *natsby.Engine) {
	if container.TestMode == true {
		e.Subscribe("insights.store.drop", natsby.WithByteReply(), container.handleDrop())
	}
	e.Subscribe("info.entry.updated", container.handleEntryUpdated())
	e.Subscribe("insights.increment.dailyCounter", natsby.WithByteReply(), container.handleIncrementDailyCounter())
	e.Subscribe("insights.get.velocity", natsby.WithByteReply(), container.handleGetVelocity())
}

func (container *Container) handleDrop() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		err := container.Repository.DropDailyCounts()
		if err != nil {
			c.Err = errors.Wrap(err, "DropDailyCounts failed")
			return
		}

		c.ByteReplyPayload = []byte("")
	}
}

func (container *Container) handleEntryUpdated() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		entryUpdatedMessage := &entryMessage.InfoEntryUpdated{}
		err := proto.Unmarshal(c.Msg.Data, entryUpdatedMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "Unmarshall failed")
			return
		}

		normalizedDay := domain.NormalizeTime(time.Unix(entryUpdatedMessage.Payload.UpdatedAt.Seconds, 0))
		day := mappers.TimeToProtoTime(normalizedDay)

		incrementDailyCounterRequest := &insightsMessage.IncrementDailyCounter{
			Payload: &insightsMessage.IncrementDailyCounter_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(incrementDailyCounterRequest)

		if err != nil {
			c.Err = errors.Wrap(err, "Marshal failed")
			return
		}

		c.NatsConnection.Publish("insights.increment.dailyCounter", request)
	}
}

func (container *Container) handleIncrementDailyCounter() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(c.Msg.Data, requestMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "unmarshall failed")
			return
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = container.Repository.IncrementDailyCounter(day, requestMessage.Payload.CreatorId)
		if err != nil {
			c.Err = errors.Wrap(err, "increment failed")
			return
		}
	}
}

func (container *Container) handleGetVelocity() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(c.Msg.Data, requestMessage)
		if err != nil {
			// TODO: Respond with error type
			c.Err = errors.Wrap(err, "unmarshall failed")
			return
		}

		normalizedStart := domain.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := domain.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyCounts, err := container.Repository.GetDailyCounts(requestMessage.Payload.CreatorId, normalizedStart, normalizedEnd)
		if err != nil {
			c.Err = errors.Wrap(err, "get daily counts failed")
			return
		}

		dailyVelocities := dailyCounts.ToVelocityScores()
		expandedDailyVelocities := domain.ExpandVelicityScores(dailyVelocities, normalizedStart, normalizedEnd)

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: mappers.VelocitiesToProtoPayload(expandedDailyVelocities),
		}

		message, err := proto.Marshal(responseMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "Marshal failed")
			return
		}

		c.ByteReplyPayload = message
	}
}

// Recovery custom recovery function
func (container *Container) Recovery() natsby.RecoveryFunc {
	return func(c *natsby.Context, err interface{}) {
		container.Logger.Error().
			Str("subject", c.Msg.Subject).
			Str("replyChan", c.Msg.Reply).
			Msg(fmt.Sprintf("%+v", err))

		if c.Msg.Reply != "" {
			errorMessage := &errorMessage.Error{}
			message, _ := proto.Marshal(errorMessage)
			c.NatsConnection.Publish(c.Msg.Reply, message)
		}
	}
}
