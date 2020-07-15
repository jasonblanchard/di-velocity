package container

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

// Container holds all service dependencies
type Container struct {
	Store           *sql.DB
	Broker          *nats.Conn
	Logger          *zerolog.Logger
	BrokerQueueName string
	TestMode        bool
}

// Input arg for Service
type Input struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDbName   string
	NatsURL          string
	Debug            bool
	Pretty           bool
	TestMode         bool
}

// New initializes service dependencies
func New(input *Input) (*Container, error) {
	container := &Container{
		BrokerQueueName: "velocity",
		TestMode:        input.TestMode,
	}

	db, err := InitalizePostgres(input.PostgresUser, input.PostgresPassword, input.PostgresDbName)
	if err != nil {
		return container, err
	}

	container.Store = db

	nc, err := InitializeNATS(input.NatsURL)
	if err != nil {
		return container, err
	}

	container.Broker = nc

	logger := InitializeLogger(input.Debug, input.Pretty)
	container.Logger = logger

	return container, nil
}

// InitializeLogger sets up the logger
func InitializeLogger(debug, pretty bool) *zerolog.Logger {
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

// InitalizePostgres sets up DB
func InitalizePostgres(user, password, dbname string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(25)
	// db.SetConnMaxLifetime(5 * time.Minute)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// InitializeNATS sets up NATS connection
func InitializeNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return nc, nil
}
