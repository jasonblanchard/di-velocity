package app

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

// Service holds all service dependencies
type Service struct {
	Store           *sql.DB
	Broker          *nats.Conn
	Logger          *zerolog.Logger
	BrokerQueueName string
	TestMode        bool
}

// ServiceInput arg for Service
type ServiceInput struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDbName   string
	NatsURL          string
	Debug            bool
	Pretty           bool
	TestMode         bool
}

// NewService initializes service dependencies
func NewService(input *ServiceInput) (Service, error) {
	service := Service{
		BrokerQueueName: "velocity",
		TestMode:        input.TestMode,
	}
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", input.PostgresUser, input.PostgresPassword, input.PostgresDbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return service, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	service.Store = db

	nc, err := nats.Connect(input.NatsURL)
	if err != nil {
		return service, err
	}

	service.Broker = nc

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if input.Pretty == true {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if input.Debug == true {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.DurationFieldUnit = time.Second
	service.Logger = &logger

	return service, nil
}

// MsgHandler message handler function. Meant to be chained with middleware.
type MsgHandler func(*nats.Msg) ([]byte, error)

// MessageHandlerChainToNatsHandler takes message handler and makes it compatible with Nats
func (service *Service) MessageHandlerChainToNatsHandler(handler MsgHandler) nats.MsgHandler {
	return func(m *nats.Msg) {
		handler(m)
	}
}

// RegisterHandler wraps handler in default middleware and listens on Nats queue
func (service *Service) RegisterHandler(topic string, handler MsgHandler) {
	// TODO: Make default middleware configurable and map over them to create this root handler
	_, handler = service.WithLogger(topic, handler)
	wrappedHandler := service.MessageHandlerChainToNatsHandler(handler)
	service.Broker.QueueSubscribe(topic, service.BrokerQueueName, wrappedHandler)
}